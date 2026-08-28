# Git Workflow Demo: Adding Signup Feature

This README explains how to demonstrate Git branching and merging using the new signup feature.

## 📋 Current State (Main Branch)

The main branch currently has:
- ✅ Login functionality
- ✅ Dashboard
- ✅ Transfer functionality
- ✅ Signup link on login page (but it doesn't work yet!)
- ❌ **NO signup route registered** (clicking "Sign up here" gives a 404 error)

## 🎯 Demo Workflow for Students

### Step 1: Show Current State
```bash
# Start the server on main branch
go run main.go signup.go

# Open browser to http://localhost:8080
# Click "Sign up here" -> 404 error (route not found)
```

**Point to make**: The signup files exist (`signup.go` and `templates/signup.html`), but the route isn't registered in `main.go`, so it doesn't work yet.

---

### Step 2: Create Feature Branch
```bash
# Create and switch to a new branch for the signup feature
git checkout -b feature/add-signup

# Show students that you're now on the feature branch
git branch
```

---

### Step 3: Activate the Signup Feature
To activate the feature, you need to register the `/signup` route in `main.go`.

**In `main.go`, find this section** (around line 300):
```go
mux := http.NewServeMux()
mux.HandleFunc("/", loginHandler)
mux.HandleFunc("/dashboard", dashboardHandler)
mux.HandleFunc("/transfer", transferHandler)
mux.HandleFunc("/lookup", lookupHandler)
mux.HandleFunc("/logout", logoutHandler)
```

**Add the signup route**:
```go
mux := http.NewServeMux()
mux.HandleFunc("/", loginHandler)
mux.HandleFunc("/dashboard", dashboardHandler)
mux.HandleFunc("/transfer", transferHandler)
mux.HandleFunc("/lookup", lookupHandler)
mux.HandleFunc("/logout", logoutHandler)
mux.HandleFunc("/signup", signupHandler)  // ← ADD THIS LINE
```

**Save the file.**

---

### Step 4: Test the New Feature
```bash
# Run the server with both main.go and signup.go
go run main.go signup.go

# Open browser to http://localhost:8080
# Click "Sign up here" -> Now it works! 🎉
```

**Try signing up with demo data:**
- First Name: `Alice`
- Last Name: `Johnson`
- Username: `alicejohnson`
- Password: `secure123`
- Confirm Password: `secure123`

After successful signup, login with the new credentials!

---

### Step 5: Commit and Merge
```bash
# Stage the changes
git add main.go

# Commit with a descriptive message
git commit -m "Add signup feature route registration"

# Switch back to main branch
git checkout main

# Merge the feature branch
git merge feature/add-signup

# Delete the feature branch (optional)
git branch -d feature/add-signup
```

---

### Step 6: Demonstrate the Effect of the Merge
```bash
# Run the server from the main branch
go run main.go signup.go

# Open browser to http://localhost:8080
# Click "Sign up here" -> It works on main branch now! ✨
```

---

## 🎓 Key Learning Points for Students

1. **Isolated Development**: Feature branches let you work on new features without breaking the main codebase
2. **404 vs Working**: Before merge, clicking signup gives 404. After merge, it works perfectly
3. **Main Branch Protected**: The main branch was never broken during development
4. **Clean Integration**: When ready, merge brings in the complete feature
5. **Code Modularity**: `signup.go` is a separate file, showing good code organization

---

## 🚀 What the Signup Feature Does

- **Validation**: Checks all fields, username length (min 3), password length (min 6), password match
- **Demo Data**: Creates new users with $10,000 starting balance
- **Success Page**: Shows a beautiful success message after registration
- **Auto-login Ready**: New users can immediately login with their credentials
- **Account Number**: Automatically generates a unique 10-digit account number

---

## 🔄 Alternative Demo: Show Conflict Resolution

If you want to demonstrate merge conflicts:

1. On main branch, add a comment in `main.go` near the routes
2. On feature branch, add a different comment in the same location
3. Try to merge - Git will show a conflict
4. Resolve the conflict manually
5. Complete the merge

This teaches students how to handle merge conflicts in real-world scenarios.

---

## 💡 Tips for the Demo

- **Keep main.go open** in your editor during the demo so students can see the one-line change
- **Use git log** to show the commit history before and after merge
- **Use git diff** to show exactly what changed between branches
- **Emphasize**: In real projects, you'd have many more changes, but the concept is the same

---

Happy teaching! 🎉
