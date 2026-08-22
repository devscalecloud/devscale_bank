package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// User holds demo account information
type User struct {
	Username  string
	Password  string
	FirstName string
	LastName  string
	Balance   float64
	AccountNo string
}

// Session holds an authenticated session entry
type Session struct {
	Username  string
	ExpiresAt time.Time
}

// DashboardData is passed to the dashboard template
type DashboardData struct {
	User             User
	BalanceFormatted string
	TransferSuccess  bool
	TransferAmount   string
	TransferTo       string
}

var (
	// Hardcoded demo users — in production use a database with bcrypt-hashed passwords
	users = map[string]User{
		"dapo": {
			Username:  "dapo",
			Password:  "dapo123",
			FirstName: "Dapo",
			LastName:  "",
			Balance:   25430.50,
			AccountNo: "1234567890",
		},
		"janesmith": {
			Username:  "janesmith",
			Password:  "secure456",
			FirstName: "Jane",
			LastName:  "Smith",
			Balance:   45820.75,
			AccountNo: "0987654321",
		},
	}

	// Demo recipient accounts for the transfer lookup feature
	demoAccounts = map[string]string{
		"0987654321": "Jane Smith",
		"1122334455": "Iyiola Olakunle Olapeju",
		"5566778899": "Adejuwon Adesanya",
		"9988776655": "Abayomi Akinyemi",
		"3344556677": "Chimezie Ogbu",
		"7766554433": "Faboya Fisayo",
	}

	sessions = map[string]Session{}
	mu       sync.RWMutex
	tmpl     *template.Template

	// appLog records banking events (login, logout, transfer, signup) to app.log and stdout
	appLog *log.Logger
)

// clientIP extracts the requester's address, preferring X-Forwarded-For if set by a proxy
func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return strings.TrimSpace(strings.Split(fwd, ",")[0])
	}
	return r.RemoteAddr
}

// generateToken creates a cryptographically secure random session token
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// formatCurrency formats a float64 as a US dollar string e.g. $25,430.50
func formatCurrency(amount float64) string {
	negative := amount < 0
	if negative {
		amount = -amount
	}
	s := fmt.Sprintf("%.2f", amount)
	parts := strings.SplitN(s, ".", 2)
	intPart := parts[0]
	decimals := parts[1]

	var result []byte
	n := len(intPart)
	for i := 0; i < n; i++ {
		if i > 0 && (n-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, intPart[i])
	}
	prefix := "$"
	if negative {
		prefix = "-$"
	}
	return prefix + string(result) + "." + decimals
}

// initials returns up to two uppercase initials from a first and last name
func initials(firstName, lastName string) string {
	var b strings.Builder
	if len(firstName) > 0 {
		b.WriteString(strings.ToUpper(firstName[:1]))
	}
	if len(lastName) > 0 {
		b.WriteString(strings.ToUpper(lastName[:1]))
	}
	return b.String()
}

// fullName joins a first and last name, omitting a blank last name
func fullName(firstName, lastName string) string {
	if lastName == "" {
		return firstName
	}
	return firstName + " " + lastName
}

// getSession retrieves the authenticated user from the session cookie
func getSession(r *http.Request) (User, bool) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return User{}, false
	}
	mu.RLock()
	session, ok := sessions[cookie.Value]
	mu.RUnlock()
	if !ok || time.Now().After(session.ExpiresAt) {
		return User{}, false
	}
	user, exists := users[session.Username]
	return user, exists
}

// loginHandler handles GET / (render login) and POST / (authenticate)
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if _, ok := getSession(r); ok {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(w, "login.html", nil); err != nil {
			log.Printf("Login template error: %v", err)
		}
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	user, ok := users[username]
	// Use constant-time comparison to prevent timing attacks
	if !ok || subtle.ConstantTimeCompare([]byte(user.Password), []byte(password)) != 1 {
		appLog.Printf("LOGIN_FAILED username=%q ip=%s", username, clientIP(r))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.ExecuteTemplate(w, "login.html", map[string]interface{}{
			"Error": "Invalid username or password. Please try again.",
		})
		return
	}

	token, err := generateToken()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	mu.Lock()
	sessions[token] = Session{
		Username:  username,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	mu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})

	appLog.Printf("LOGIN_SUCCESS username=%q ip=%s", username, clientIP(r))

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// dashboardHandler renders the main banking dashboard
func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := DashboardData{
		User:             user,
		BalanceFormatted: formatCurrency(user.Balance),
		TransferSuccess:  r.URL.Query().Get("success") == "1",
		TransferAmount:   r.URL.Query().Get("amount"),
		TransferTo:       r.URL.Query().Get("to"),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "dashboard.html", data); err != nil {
		log.Printf("Dashboard template error: %v", err)
	}
}

// transferHandler processes a demo money transfer (POST /transfer)
func transferHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := getSession(r)
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	amount := r.FormValue("amount")
	accountNo := r.FormValue("account_number")

	recipientName := demoAccounts[accountNo]
	if recipientName == "" {
		recipientName = "the recipient"
	}

	appLog.Printf("TRANSFER username=%q from_account=%s to_account=%s to_name=%q amount=%s ip=%s",
		user.Username, user.AccountNo, accountNo, recipientName, amount, clientIP(r))

	// PRG pattern: redirect after POST to prevent form resubmission on refresh
	redirectURL := "/dashboard?success=1&amount=" + url.QueryEscape(amount) + "&to=" + url.QueryEscape(recipientName)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// lookupHandler returns recipient name for a 10-digit account number (GET /lookup)
func lookupHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := getSession(r); !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	accountNo := r.URL.Query().Get("account")

	// Validate: exactly 10 digits only
	if len(accountNo) != 10 {
		http.Error(w, "Account number must be 10 digits", http.StatusBadRequest)
		return
	}
	for _, c := range accountNo {
		if c < '0' || c > '9' {
			http.Error(w, "Account number must contain digits only", http.StatusBadRequest)
			return
		}
	}

	name, ok := demoAccounts[accountNo]
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, name)
}

// logoutHandler destroys the session and redirects to login
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		mu.Lock()
		session, ok := sessions[cookie.Value]
		delete(sessions, cookie.Value)
		mu.Unlock()
		if ok {
			appLog.Printf("LOGOUT username=%q ip=%s", session.Username, clientIP(r))
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	appLog = log.New(os.Stdout, "", log.Ldate|log.Ltime)

	var err error
	tmpl, err = template.New("").Funcs(template.FuncMap{
		"initials": initials,
		"fullName": fullName,
	}).ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v\nMake sure you run this from the project root directory.", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", loginHandler)
	mux.HandleFunc("/dashboard", dashboardHandler)
	mux.HandleFunc("/transfer", transferHandler)
	mux.HandleFunc("/lookup", lookupHandler)
	mux.HandleFunc("/logout", logoutHandler)
	mux.HandleFunc("/signup", signupHandler)

	port := ":8080"
	fmt.Println("============================================")
	fmt.Println("  DevScale Banking App")
	fmt.Println("============================================")
	fmt.Printf("  Server running at http://0.0.0.0%s\n", port)
	fmt.Println("  Open your browser: http://localhost:8080")
	fmt.Println("  Demo login: dapo / dapo123")
	fmt.Println("============================================")
	log.Fatal(http.ListenAndServe(port, mux))
}
