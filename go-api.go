package main

import (
	"context"
	"crypto/md5"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
	"websocket/websocket"

	"github.com/gorilla/mux"
)

func wrapJwt(
	jwt *JWTService,
	f func(http.ResponseWriter, *http.Request, *JWTService),
) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		f(rw, r, jwt)
	}
}

func (uServ *UserService) addAdmin() error {
	CAKE_ADMIN_EMAIL := os.Getenv("CAKE_ADMIN_EMAIL")
	CAKE_ADMIN_PASSWORD := os.Getenv("CAKE_ADMIN_PASSWORD")
	passwordDigest := md5.New().Sum([]byte(CAKE_ADMIN_PASSWORD))
	admin := User{
		Email:          CAKE_ADMIN_EMAIL,
		PasswordDigest: string(passwordDigest),
		FavoriteCake:   "AdminCake",
		Role:           "AdminRole",
		Ban:            false,
		BanHistory:     History{},
	}
	err := uServ.repository.Add(admin.Email, admin)
	if err != nil {
		return err
	}
	return nil
}

var JwtService, jwt_err = NewJWTService("pubkey.rsa", "privkey.rsa")

func main() {
	if jwt_err != nil {
		panic(jwt_err)
	}

	os.Setenv("CAKE_ADMIN_EMAIL", "admin@mail.com")
	os.Setenv("CAKE_ADMIN_PASSWORD", "adminadmin")

	r := mux.NewRouter()

	Hub := websocket.NewHub()
	go Hub.Run()

	users := NewInMemoryUserStorage()
	userService := UserService{
		repository: users,
	}

	r.HandleFunc("/admin/ban", logRequest(JwtService.jwtAuthAdmin(userService.repository, measureResponseDuration(banUserHandler)))).Methods(http.MethodPost)
	r.HandleFunc("/admin/unban", logRequest(JwtService.jwtAuthAdmin(userService.repository, measureResponseDuration(unbanUserHandler)))).Methods(http.MethodPost)
	r.HandleFunc("/admin/inspect", logRequest(JwtService.jwtAuthAdmin(userService.repository, measureResponseDuration(inspectHandler)))).Methods(http.MethodGet)

	r.HandleFunc("/user/register", logRequest(userService.Register)).Methods(http.MethodPost)
	r.HandleFunc("/user/jwt", logRequest(wrapJwt(JwtService, userService.JWT))).Methods(http.MethodPost)

	r.HandleFunc("/user/me", logRequest(JwtService.jwtAuth(users, measureResponseDuration(getMeHandler)))).Methods(http.MethodGet)
	r.HandleFunc("/user/my_cake", logRequest(JwtService.jwtAuth(users, measureResponseDuration(getCakeHandler)))).Methods(http.MethodGet)
	r.HandleFunc("/user/my_email", logRequest(JwtService.jwtAuth(users, measureResponseDuration(getEmailHandler)))).Methods(http.MethodGet)

	r.HandleFunc("/user/favorite_cake", logRequest(JwtService.jwtAuth(users, measureResponseDuration(userService.updateCakeHandler)))).Methods(http.MethodPut)
	r.HandleFunc("/user/email", logRequest(JwtService.jwtAuth(users, measureResponseDuration(userService.updateEmailHandler)))).Methods(http.MethodPut)
	r.HandleFunc("/user/password", logRequest(JwtService.jwtAuth(users, measureResponseDuration(userService.updatePasswordHandler)))).Methods(http.MethodPut)
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(Hub, w, r)
	})

	userService.addAdmin()
	srv := http.Server{
		Addr:    ":8085",
		Handler: r,
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go Run()
	go websocket.PrometheusRun()

	go func() {
		<-interrupt
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	log.Printf("Server started, press cntrl + C to stop ")
	err := srv.ListenAndServe()
	if err != nil {
		log.Println("Server exited with error:", err)
	}
	log.Println("Good bye :)")
}
