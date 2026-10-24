package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/anshggss/goservers/internal/auth"
	"github.com/anshggss/goservers/internal/database"
	"github.com/google/uuid"
)

type apiConfig struct {
	serverRequests atomic.Int32
	db             *database.Queries
	secret         string
}

func (a *apiConfig) middlewareMetrics(next http.Handler) http.Handler {
	// here we need to return a function and we can't simply write the handler's functionality in the middleware because the middlewareMetrics gets called during registration of the server and later on the returned function gets registered so we need to wrap the original next function inside another function that will do the middleware things
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.serverRequests.Add(1)
		fmt.Printf("Current server requests: %v\n", a.serverRequests.Load())
		next.ServeHTTP(w, r)
	})
}

func health(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	io.WriteString(w, "OK")
}

func (a *apiConfig) metrics(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	io.WriteString(w, fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, a.serverRequests.Load()))
}

func (a *apiConfig) reset(w http.ResponseWriter, req *http.Request) {

	defer req.Body.Close()
	platform := os.Getenv("PLATFORM")
	if platform != "dev" {
		replyWithError(w, 403, "Forbidden")
		return
	}
	err := a.db.DeleteUsers(req.Context())
	if err != nil {
		replyWithError(w, 400, "Issue deleting the users")
		return
	}

	w.WriteHeader(200)
	a.serverRequests = atomic.Int32{}

}

func replyWithError(w http.ResponseWriter, code int, msg string) {

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	payload := struct {
		Error string `json:"error"`
	}{
		Error: msg,
	}
	formatted, err := json.Marshal(payload)
	if err != nil {
		return
	}
	w.Write(formatted)
}

func replyWithJson(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	pl, err := json.Marshal(payload)
	if err != nil {
		replyWithError(w, 400, "Something went wrong")
		return
	}
	w.Write(pl)
}

func cleanBody(msg string) string {

	notAllowed := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}
	var res []string
	for _, word := range strings.Fields(msg) {
		_, ok := notAllowed[strings.ToLower(word)]
		if ok {
			res = append(res, strings.Repeat("*", 4))
			continue
		}
		res = append(res, word)

	}
	return strings.Join(res, " ")
}

func (a *apiConfig) createChirps(w http.ResponseWriter, r *http.Request) {

	type rB struct {
		Body   string    `json:"body"`
		Userid uuid.UUID `json:"user_id"`
	}
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	var req rB
	if err := decoder.Decode(&req); err != nil {
		replyWithError(w, 400, "Something went wrong")
		return
	}
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		replyWithError(w, 400, err.Error())
	}
	uid, err := auth.ValidateJWT(token, a.secret)
	if err != nil {
		replyWithError(w, 401, "Unauthorized")
		return
	}

	if len(req.Body) > 140 {
		replyWithError(w, 400, "Chirp is too long")
		return
	}
	cleaned := cleanBody(req.Body)
	dbchirp := database.CreateChirpParams{
		ID:   uuid.New(),
		Body: cleaned,
		UserID: uuid.NullUUID{
			UUID:  uid,
			Valid: true,
		},
	}
	rC, err := a.db.CreateChirp(r.Context(), dbchirp)
	if err != nil {
		replyWithError(w, 400, fmt.Sprintf("An Error occured while adding your chirp: %s", err.Error()))
		return
	}

	returningChirp := chirp{
		ID:        rC.ID,
		CreatedAt: rC.CreatedAt,
		UpdatedAt: rC.UpdatedAt,
		Body:      rC.Body,
		User_id:   rC.UserID.UUID,
	}
	replyWithJson(w, 201, returningChirp)

}

type chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	User_id   uuid.UUID `json:"user_id"`
}

func (a *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	dbChirps, err := a.db.GetChirps(r.Context())
	if err != nil {
		replyWithError(w, 401, fmt.Sprintf("An error occured : %s\n", err.Error()))
		return
	}

	var chirps []chirp
	for _, cp := range dbChirps {
		chirps = append(chirps, chirp{
			ID:        cp.ID,
			CreatedAt: cp.CreatedAt,
			UpdatedAt: cp.UpdatedAt,
			Body:      cp.Body,
			User_id:   cp.UserID.UUID,
		})
	}
	replyWithJson(w, 200, chirps)
}
func (a *apiConfig) getChirp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	chirpId, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		replyWithError(w, 404, "Chirp ID not found")
		return
	}
	dbChirp, err := a.db.GetChirpbyID(r.Context(), chirpId)
	if err != nil {
		replyWithError(w, 404, "Chirp ID not found")
		return
	}
	tchirp := chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		User_id:   dbChirp.UserID.UUID,
	}
	replyWithJson(w, 200, tchirp)
}
func (a *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	type user struct {
		Email    string
		Password string
	}
	var usr user
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if err := decoder.Decode(&usr); err != nil {
		replyWithError(w, 400, "Error decoding the request's body")
		return
	}
	hashed, err := auth.HashPassword(usr.Password)
	if err != nil {
		replyWithError(w, 400, "Error parsing the password")
	}

	dbUser, err := a.db.CreateUser(r.Context(), database.CreateUserParams{
		ID:             uuid.New(),
		Email:          usr.Email,
		HashedPassword: hashed,
	})
	if err != nil {
		replyWithError(w, 400, fmt.Sprintf("Error adding user to the database: %s", err.Error()))
		return
	}

	ndbUser := struct {
		ID        uuid.UUID    `json:"id"`
		CreatedAt sql.NullTime `json:"created_at"`
		UpdatedAt sql.NullTime `json:"updated_at"`
		Email     string       `json:"email"`
	}{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}

	replyWithJson(w, 201, ndbUser)

}

func (a *apiConfig) loginUser(w http.ResponseWriter, r *http.Request) {
	type user struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req user
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if err := decoder.Decode(&req); err != nil {
		replyWithError(w, 400, "Error decoding the request's body")
		return
	}
	isPresentUser, err := a.db.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		replyWithError(w, 401, "Unauthorized")
		return
	}
	isValid, err := auth.CheckPasswordHash(req.Password, isPresentUser.HashedPassword)
	if err != nil || !isValid {
		replyWithError(w, 401, "Unauthorized")
		return
	}
	expiryAccess := "1h"

	dur, err := time.ParseDuration(expiryAccess)
	if err != nil {
		replyWithError(w, 401, "Bad timing")
		return
	}

	refToken := auth.MakeRefreshToken()
	err = a.db.CreateRefresh(r.Context(), database.CreateRefreshParams{
		Token:  refToken,
		UserID: isPresentUser.ID,
		ExpiresAt: sql.NullTime{
			Time:  time.Now().Add(time.Hour * 24 * 60),
			Valid: true,
		},
	})

	token, err := auth.MakeJWT(isPresentUser.ID, a.secret, dur)

	ndbUser := struct {
		ID           uuid.UUID    `json:"id"`
		CreatedAt    sql.NullTime `json:"created_at"`
		UpdatedAt    sql.NullTime `json:"updated_at"`
		Email        string       `json:"email"`
		Token        string       `json:"token"`
		RefreshToken string       `json:"refresh_token"`
	}{
		ID:           isPresentUser.ID,
		CreatedAt:    isPresentUser.CreatedAt,
		UpdatedAt:    isPresentUser.UpdatedAt,
		Email:        isPresentUser.Email,
		Token:        token,
		RefreshToken: refToken,
	}

	replyWithJson(w, 200, ndbUser)

}

func (a *apiConfig) createNewToken(w http.ResponseWriter, r *http.Request) {
	refToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		replyWithError(w, 400, err.Error())
		return
	}
	dbRefToken, err := a.db.GetToken(r.Context(), refToken)
	// Refresh token doesn't exist
	if err != nil || time.Now().Compare(dbRefToken.ExpiresAt.Time) > 0 || dbRefToken.RevokedAt.Valid {
		replyWithError(w, 401, "Unauthorized")
		return
	}
	accessToken, err := auth.MakeJWT(dbRefToken.UserID, a.secret, time.Hour)
	if err != nil {
		replyWithError(w, 500, "We can't even make an access token rn pls BEAR with us :)")
		return
	}
	replyWithJson(w, 200, struct {
		Token string `json:"token"`
	}{
		Token: accessToken,
	})
}
func (a *apiConfig) revokeToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		replyWithError(w, 400, err.Error())
		return
	}
	revokeTime := time.Now()
	err = a.db.RevokeToken(r.Context(), database.RevokeTokenParams{
		UpdatedAt: sql.NullTime{
			Time:  revokeTime,
			Valid: true,
		},
		RevokedAt: sql.NullTime{
			Time:  revokeTime,
			Valid: true,
		},
		Token: refreshToken,
	})
	if err != nil {
		replyWithError(w, 400, err.Error())
		return
	}
	replyWithJson(w, 204, "")
}
func (a *apiConfig) updateUser(w http.ResponseWriter, r *http.Request) {
	var usr struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	defer r.Body.Close()
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		replyWithError(w, 401, err.Error())
		return
	}

	uid, err := auth.ValidateJWT(tokenString, a.secret)
	if err != nil {
		replyWithError(w, 401, err.Error())
		return
	}

	fmt.Printf("%s\n", usr.Email)
	curUser, err := a.db.GetUserbyID(r.Context(), uid)
	if err != nil {
		replyWithError(w, 400, err.Error())
		return
	}
	if uid != curUser.ID {
		replyWithError(w, 401, "Unauthorized")
		return
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&usr); err != nil {
		replyWithError(w, 400, "Bad body")
		return
	}
	hashedPassword, err := auth.HashPassword(usr.Password)
	if err != nil {
		replyWithError(w, 401, "Unauthorized")
		return
	}
	nUsr, err := a.db.UpdateUser(r.Context(), database.UpdateUserParams{
		Email:          usr.Email,
		ID:             curUser.ID,
		HashedPassword: hashedPassword,
		UpdatedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
	})
	payload := struct {
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		ID        uuid.UUID `json:"id"`
	}{
		ID:        nUsr.ID,
		Email:     nUsr.Email,
		CreatedAt: nUsr.CreatedAt.Time,
		UpdatedAt: nUsr.UpdatedAt.Time,
	}
	replyWithJson(w, 200, payload)

}

func (a *apiConfig) deleteChirp(w http.ResponseWriter, r *http.Request) {
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		replyWithError(w, 401, err.Error())
		return
	}
	uid, err := auth.ValidateJWT(tokenString, a.secret)
	if err != nil {
		replyWithError(w, 400, err.Error())
		return
	}
	chirpId, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {

		replyWithError(w, 400, err.Error())
		return
	}
	dbChirp, err := a.db.GetChirpbyID(r.Context(), chirpId)
	if err != nil {

		replyWithError(w, 404, err.Error())
		return
	}
	if dbChirp.UserID.UUID != uid {
		replyWithError(w, 403, "Unauthorized")
		return
	}
	err = a.db.DeleteChirpByID(r.Context(), chirpId)
	if err != nil {
		replyWithError(w, 400, err.Error())
		return
	}
	replyWithJson(w, 204, "")
}

func (a *apiConfig) upgradeUser(w http.ResponseWriter, r *http.Request) {

}
