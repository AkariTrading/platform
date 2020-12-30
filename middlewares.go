package main

import (
	"net/http"
)

func authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		return

		// logger := middleware.GetLogger(r)

		// DEBUG
		// ctx := context.WithValue(r.Context(), middleware.USERID, "d736b408-aa60-43a3-8daa-d6c21a23c417")
		// next.ServeHTTP(w, r.WithContext(ctx))
		// return
		// DEBUG

		// var sessionToken string

		// if token := r.Header.Get(sessionTokenHeader); token != "" {
		// 	sessionToken = token
		// } else if c, err := r.Cookie("session_token"); err == nil {

		// 	fmt.Println(c)
		// 	sessionToken = c.Value
		// }

		// if sessionToken == "" {
		// 	w.WriteHeader(http.StatusUnauthorized)
		// 	return
		// }

		// response, err := redis.String(redisHandle.Do(redis.GetKey, sessionToken))

		// if err != nil {
		// 	logger.Error(err)
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	return
		// }
		// if response == "" {
		// 	w.WriteHeader(http.StatusUnauthorized)
		// 	return
		// }

		// next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), middleware.USERID, response)))
	})
}

func jsonResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if (*r).Method == "OPTIONS" {
			return
		}
		next.ServeHTTP(w, r)
	})
}
