package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// SignupData is passed to the signup template
type SignupData struct {
	Error   string
	Success bool
}

// signupHandler handles both GET /signup (render form) and POST /signup (process registration)
func signupHandler(w http.ResponseWriter, r *http.Request) {
	// If user is already logged in, redirect to dashboard
	if _, ok := getSession(r); ok {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(w, "signup.html", SignupData{}); err != nil {
			log.Printf("Signup template error: %v", err)
		}
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")
	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")

	// Validation
	if username == "" || password == "" || firstName == "" || lastName == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.ExecuteTemplate(w, "signup.html", SignupData{
			Error: "All fields are required.",
		})
		return
	}

	if len(username) < 3 {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.ExecuteTemplate(w, "signup.html", SignupData{
			Error: "Username must be at least 3 characters long.",
		})
		return
	}

	if len(password) < 6 {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.ExecuteTemplate(w, "signup.html", SignupData{
			Error: "Password must be at least 6 characters long.",
		})
		return
	}

	if password != confirmPassword {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.ExecuteTemplate(w, "signup.html", SignupData{
			Error: "Passwords do not match.",
		})
		return
	}

	// Check if username already exists
	mu.RLock()
	_, exists := users[username]
	mu.RUnlock()

	if exists {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.ExecuteTemplate(w, "signup.html", SignupData{
			Error: "Username already exists. Please choose a different username.",
		})
		return
	}

	// Generate a random account number (10 digits)
	accountNo := fmt.Sprintf("%010d", time.Now().UnixNano()%10000000000)

	// Create new user with starting balance
	newUser := User{
		Username:  username,
		Password:  password, // In production, hash this with bcrypt!
		FirstName: firstName,
		LastName:  lastName,
		Balance:   10000.00, // Starting balance: $10,000
		AccountNo: accountNo,
	}

	// Add user to the users map
	mu.Lock()
	users[username] = newUser
	// Also add to demo accounts for transfers
	demoAccounts[accountNo] = firstName + " " + lastName
	mu.Unlock()

	appLog.Printf("SIGNUP username=%q account=%s ip=%s", username, accountNo, clientIP(r))

	// Show success page
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.ExecuteTemplate(w, "signup.html", SignupData{
		Success: true,
	})
}
