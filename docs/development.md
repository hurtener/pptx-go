# Development Guide

---

## Local Development Setup

When developing locally, you need to use the local code instead of the remote repository. There are several ways to achieve this:

### Method 1: Using `go.mod` replace directive (Recommended)

Add a `replace` directive to your `go.mod`:

```go
module github.com/hurtener/pptx-go

go 1.24.5

// For local development, point to local directory
replace github.com/hurtener/pptx-go => /path/to/your/local/Go-pptx
```

This tells Go to use the local directory instead of fetching from GitHub.

### Method 2: Using `go.mod` with relative path

If your project using this library is in a sibling directory:

```go
replace github.com/hurtener/pptx-go => ../Go-pptx
```

### Method 3: Using `go.work` (Go 1.18+)

Create a `go.work` file in your workspace root:

```go
go 1.24.5

use (
    ./Go-pptx           // This library
    ./my-pptx-project   // Your project using this library
)
```

### Method 4: Using GOPATH/src (Traditional)

```bash
# Clone to GOPATH
mkdir -p $GOPATH/src/github.com/hurtener
cd $GOPATH/src/github.com/hurtener
git clone https://github.com/hurtener/pptx-go.git
```

### Windows Path Examples

```go
// Windows absolute path
replace github.com/hurtener/pptx-go => C:\Users\ASUS\mywork\Go-pptx

// Windows relative path
replace github.com/hurtener/pptx-go => ..\Go-pptx
```

### Quick Setup Script

Create a `dev_setup.sh` (or `dev_setup.bat` for Windows):

```bash
#!/bin/bash
# Run this in your project that uses go-pptx

# Add replace directive temporarily for development
go mod edit -replace github.com/hurtener/pptx-go=/path/to/Go-pptx

# Verify
go list -m github.com/hurtener/pptx-go
```

Create `dev_setup.bat` (Windows):

```batch
@echo off
REM Run this script in the project that uses go-pptx

REM Add a temporary replace directive for development
go mod edit -replace github.com/hurtener/pptx-go=C:\Users\ASUS\mywork\Go-pptx

REM Verify
go list -m github.com/hurtener/pptx-go
```

### Before Publishing

**Important:** Remove or comment out the `replace` directive before publishing your project:

```go
// replace github.com/hurtener/pptx-go => ../Go-pptx  // Uncomment for local dev
```

Or use a separate `go.mod.dev` file and swap as needed.

### Recommended Workspace Layout

```
Example project structure:
/workspace
├── Go-pptx/                    # This library (under development)
│   ├── go.mod
│   ├── opc/
│   └── ...
├── my-pptx-app/                # Application using this library
│   ├── go.mod                  # Contains replace directive
│   └── main.go
└── go.work                     # Optional: workspace file
```

**go.work example:**
```go
go 1.24.5

use (
    ./Go-pptx
    ./my-pptx-app
)
```

When using `go.work`, Go resolves local modules automatically — no `replace` directive is needed.
