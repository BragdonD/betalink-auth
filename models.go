// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package betalinkauth

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Emailverification struct {
	UserID            pgtype.UUID
	VerificationToken string
	CreatedAt         pgtype.Timestamp
	Used              bool
}

type Externalloginprovider struct {
	ProviderID       pgtype.UUID
	ProviderName     string
	ProviderEndpoint string
}

type Hashalgorithm struct {
	Hashalgorithm string
}

type Passwordrecovery struct {
	UserID        pgtype.UUID
	RecoveryToken string
	CreatedAt     pgtype.Timestamp
	Used          bool
}

type User struct {
	UserID    pgtype.UUID
	FirstName string
	LastName  string
	CreatedAt pgtype.Timestamp
	UpdatedAt pgtype.Timestamp
}

type Userloginexternal struct {
	UserID               pgtype.UUID
	ProviderID           pgtype.UUID
	ProviderAccessToken  string
	ProviderRefreshToken string
}

type Userslogindatum struct {
	UserID        pgtype.UUID
	Email         string
	Passwordhash  string
	Passwordsalt  string
	Hashalgorithm string
}
