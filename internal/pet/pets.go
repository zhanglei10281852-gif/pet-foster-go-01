package pet

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func (s *Service) AddPet(ctx context.Context, principal Principal, input Pet) (Pet, error) {
	if principal.UserID == 0 {
		return Pet{}, ErrUnauthenticated
	}
	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Type) == "" {
		return Pet{}, fmt.Errorf("%w: pet name and type are required", ErrValidation)
	}
	now := formatStoredTime(s.now())
	ownerID := input.OwnerID
	if ownerID == 0 || principal.Role != RoleAdmin {
		ownerID = principal.UserID
	}
	result, err := s.store.db.ExecContext(ctx, `INSERT INTO pets(pet_name,pet_type,breed,age,weight,health_status,special_requirements,avatar,owner_id,create_time,update_time) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, input.Name, input.Type, input.Breed, input.Age, input.Weight, input.HealthStatus, input.SpecialRequirements, input.Avatar, ownerID, now, now)
	if err != nil {
		return Pet{}, translateServiceError(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Pet{}, err
	}
	return s.GetPet(ctx, principal, id)
}

func (s *Service) UpdatePet(ctx context.Context, principal Principal, input Pet) (Pet, error) {
	if principal.UserID == 0 {
		return Pet{}, ErrUnauthenticated
	}
	if input.ID == 0 || strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Type) == "" {
		return Pet{}, fmt.Errorf("%w: pet id, name and type are required", ErrValidation)
	}
	pet, err := s.GetPet(ctx, principal, input.ID)
	if err != nil {
		return Pet{}, err
	}
	if principal.Role != RoleAdmin && pet.OwnerID != principal.UserID {
		return Pet{}, ErrForbidden
	}
	ownerID := pet.OwnerID
	if principal.Role == RoleAdmin && input.OwnerID != 0 {
		ownerID = input.OwnerID
	}
	_, err = s.store.db.ExecContext(ctx, `UPDATE pets SET pet_name=?,pet_type=?,breed=?,age=?,weight=?,health_status=?,special_requirements=?,avatar=?,owner_id=?,update_time=? WHERE pet_id=?`, input.Name, input.Type, input.Breed, input.Age, input.Weight, input.HealthStatus, input.SpecialRequirements, input.Avatar, ownerID, formatStoredTime(s.now()), input.ID)
	if err != nil {
		return Pet{}, translateServiceError(err)
	}
	return s.GetPet(ctx, principal, input.ID)
}

func (s *Service) DeletePet(ctx context.Context, principal Principal, id int64) error {
	if principal.UserID == 0 {
		return ErrUnauthenticated
	}
	pet, err := s.GetPet(ctx, principal, id)
	if err != nil {
		return err
	}
	if principal.Role != RoleAdmin && pet.OwnerID != principal.UserID {
		return ErrForbidden
	}
	result, err := s.store.db.ExecContext(ctx, `DELETE FROM pets WHERE pet_id=?`, id)
	if err != nil {
		return err
	}
	return requireAffected(result, ErrNotFound)
}

func (s *Service) GetPet(ctx context.Context, principal Principal, id int64) (Pet, error) {
	if principal.UserID == 0 {
		return Pet{}, ErrUnauthenticated
	}
	var item Pet
	var created, updated string
	err := s.store.db.QueryRowContext(ctx, `SELECT p.pet_id,p.pet_name,p.pet_type,COALESCE(p.breed,''),COALESCE(p.age,0),COALESCE(p.weight,0),COALESCE(p.health_status,''),COALESCE(p.special_requirements,''),COALESCE(p.avatar,''),p.owner_id,COALESCE(u.username,''),p.create_time,p.update_time FROM pets p LEFT JOIN pet_users u ON u.user_id=p.owner_id WHERE p.pet_id=?`, id).Scan(&item.ID, &item.Name, &item.Type, &item.Breed, &item.Age, &item.Weight, &item.HealthStatus, &item.SpecialRequirements, &item.Avatar, &item.OwnerID, &item.OwnerName, &created, &updated)
	if err == sql.ErrNoRows {
		return Pet{}, ErrNotFound
	}
	if err != nil {
		return Pet{}, err
	}
	if principal.Role != RoleAdmin && item.OwnerID != principal.UserID {
		return Pet{}, ErrForbidden
	}
	item.CreatedAt, _ = parseStoredTime(created)
	item.UpdatedAt, _ = parseStoredTime(updated)
	return item, nil
}

func (s *Service) ListPets(ctx context.Context, principal Principal, pageNum, pageSize int, name, typ string, ownerID int64) (Page[Pet], error) {
	if principal.UserID == 0 {
		return Page[Pet]{}, ErrUnauthenticated
	}
	pageNum, pageSize = normalizePage(pageNum, pageSize)
	where := []string{"1=1"}
	args := []any{}
	if principal.Role != RoleAdmin {
		where = append(where, "p.owner_id=?")
		args = append(args, principal.UserID)
	} else if ownerID != 0 {
		where = append(where, "p.owner_id=?")
		args = append(args, ownerID)
	}
	if strings.TrimSpace(name) != "" {
		where = append(where, "p.pet_name LIKE ?")
		args = append(args, "%"+strings.TrimSpace(name)+"%")
	}
	if strings.TrimSpace(typ) != "" {
		where = append(where, "p.pet_type=?")
		args = append(args, typ)
	}
	clause := strings.Join(where, " AND ")
	var total int
	if err := s.store.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pets p WHERE "+clause, args...).Scan(&total); err != nil {
		return Page[Pet]{}, err
	}
	query := "SELECT p.pet_id,p.pet_name,p.pet_type,COALESCE(p.breed,''),COALESCE(p.age,0),COALESCE(p.weight,0),COALESCE(p.health_status,''),COALESCE(p.special_requirements,''),COALESCE(p.avatar,''),p.owner_id,COALESCE(u.username,''),p.create_time,p.update_time FROM pets p LEFT JOIN pet_users u ON u.user_id=p.owner_id WHERE " + clause + " ORDER BY p.pet_id DESC LIMIT ? OFFSET ?"
	rows, err := s.store.db.QueryContext(ctx, query, append(args, pageSize, (pageNum-1)*pageSize)...)
	if err != nil {
		return Page[Pet]{}, err
	}
	defer rows.Close()
	items := make([]Pet, 0)
	for rows.Next() {
		var item Pet
		var created, updated string
		if err := rows.Scan(&item.ID, &item.Name, &item.Type, &item.Breed, &item.Age, &item.Weight, &item.HealthStatus, &item.SpecialRequirements, &item.Avatar, &item.OwnerID, &item.OwnerName, &created, &updated); err != nil {
			return Page[Pet]{}, err
		}
		item.CreatedAt, _ = parseStoredTime(created)
		item.UpdatedAt, _ = parseStoredTime(updated)
		items = append(items, item)
	}
	return Page[Pet]{List: items, Total: total, PageNum: pageNum, PageSize: pageSize}, rows.Err()
}

func (s *Service) MyPets(ctx context.Context, principal Principal) ([]Pet, error) {
	page, err := s.ListPets(ctx, principal, 1, 100, "", "", 0)
	return page.List, err
}
