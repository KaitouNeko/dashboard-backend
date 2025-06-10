package user

import (
	"ai-workshop/internal/auth"
	"ai-workshop/internal/models"
	"ai-workshop/internal/utils/errorutils"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	Repo *UserRepository
}

func NewUserService(repo *UserRepository) *UserService {
	return &UserService{
		Repo: repo,
	}
}

func (s *UserService) GetUserByIdWithPasswordService(id uuid.UUID) (*models.User, error) {
	return s.Repo.GetByIdWithPassword(id)
}
func (s *UserService) GetUsersService() (*[]models.User, error) {
	return s.Repo.GetUsers()
}

func (s *UserService) GetUserByIdService(id uuid.UUID) (*models.User, error) {
	return s.Repo.GetById(id)
}

func (s *UserService) CreateUserService(user models.User) error {
	hashedPw, err := s.HashPassword(user.Password)

	if err != nil {
		return fmt.Errorf("Error when attempting to hash password.")
	}

	// update user's password with hashed password.
	user.Password = hashedPw

	return s.Repo.Create(user)
}

func (s *UserService) UpdatePasswordUserService(requestData UserUpdatePasswordRequest, userId uuid.UUID) error {

	user, _ := s.GetUserByIdWithPasswordService(userId)

	if requestData.NewPassword != requestData.RepeatNewPassword {
		return fmt.Errorf("NewPassword and RepeatNewPassword are different")
	}

	isSame, _ := s.ComparePasswords(user.Password, requestData.Password)
	if !isSame {
		return fmt.Errorf("Stored password and input password are different")
	}

	hashedPw, err := s.HashPassword(requestData.NewPassword)
	if err != nil {
		return fmt.Errorf("Error when attempting to hash password.")
	}

	params := UserUpdatePasswordParams{
		ID:       userId,
		Password: hashedPw,
	}

	return s.Repo.UpdatePassword(params)
}

func (s *UserService) UpdateInfoUserHandler(request UserUpdateInfoRequest, userId uuid.UUID) error {
	params := UserUpdateInfoParams{
		ID:   userId,
		Name: request.Name,
		// Status: request.Status,
		Permission: request.Permission,
	}
	return s.Repo.UpdateInfo(params, userId)
}

func (s *UserService) LoginUserService(loginReq UserLoginRequest) (*UserLoginResponse, error) {
	user, err := s.Repo.GetUserByEmail(loginReq.Email)

	if err != nil {
		return nil, errors.New("Could not get user with provided email.")
	}

	// extract password, and compare hashes
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password)); err != nil {
		return nil, errorutils.ErrUnauthorized
	}

	// construct response with both user info and auth credentials
	accessExpiryTime := time.Minute * 60
	accessToken, err := auth.GenerateJWT(*user, auth.Access, accessExpiryTime)
	refreshExpiryTime := time.Hour * 24 * 7
	refreshToken, err := auth.GenerateJWT(*user, auth.Refresh, refreshExpiryTime)

	user.Password = ""

	res := &UserLoginResponse{
		AccessToken:      accessToken,
		AccessExpiresIn:  int(accessExpiryTime),
		RefreshToken:     refreshToken,
		RefreshExpiresIn: int(refreshExpiryTime),
		UserInfo:         user,
	}

	return res, nil
}

// HashPassword hashes the given password using bcrypt.
func (s *UserService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s *UserService) ComparePasswords(storedPassword string, inputPassword string) (bool, error) {
	// Compare the hashed password with the user-provided password
	err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(inputPassword))
	if err != nil {
		// If error is not nil, the passwords do not match
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil // Passwords do not match
		}
		// Handle other potential errors (e.g., hash format issues)
		return false, err
	}
	// If err is nil, the passwords match
	return true, nil
}

/**
* Create Default Users.
**/

func (s *UserService) CreateDefaultUsersService(Users []CreateDefaultUser) error {

	var hashedPwUsers []CreateDefaultUser

	// update Users passwords with hash
	for _, User := range Users {
		hashedPw, err := s.HashPassword(User.Password)

		if err != nil {
			return err
		}
		User.Password = hashedPw

		hashedPwUsers = append(hashedPwUsers, User)
	}

	return s.Repo.CreateDefaultUsers(hashedPwUsers)
}
