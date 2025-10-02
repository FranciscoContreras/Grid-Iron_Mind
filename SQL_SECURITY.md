# SQL Injection Prevention & Security

## Overview

This document outlines SQL security practices implemented in Grid Iron Mind API and verification procedures to ensure the codebase remains safe from SQL injection attacks.

## Current Security Status

✅ **ALL database queries use parameterized statements**
✅ **NO direct string concatenation of user input in SQL**
✅ **pgx/v5 driver with prepared statement support**
✅ **Input validation before database queries**

## Parameterized Query Pattern

### ✅ SAFE - Using Placeholders

All database queries follow this safe pattern:

```go
// SAFE: Using $1, $2, $3 placeholders
query := `
    SELECT id, name, position
    FROM players
    WHERE position = $1 AND status = $2
`
rows, err := pool.Query(ctx, query, position, status)
```

### ❌ UNSAFE - String Concatenation (NOT USED IN CODEBASE)

This pattern is NOT used anywhere in the codebase:

```go
// UNSAFE: DO NOT DO THIS
query := "SELECT * FROM players WHERE name = '" + userInput + "'"
rows, err := pool.Query(ctx, query)  // VULNERABLE TO SQL INJECTION
```

## Verified Safe Query Patterns

### 1. Static WHERE Clauses
```go
// Building dynamic WHERE clauses safely
whereClause := " WHERE 1=1"
args := []interface{}{}
argCount := 1

if position != "" {
    whereClause += fmt.Sprintf(" AND position = $%d", argCount)
    args = append(args, position)
    argCount++
}

query := "SELECT * FROM players" + whereClause
rows, err := pool.Query(ctx, query, args...)
```

**Why This Is Safe:**
- Only SQL keywords and placeholder numbers are concatenated
- Actual user input values are passed separately via `args...`
- `fmt.Sprintf` only adds `$1, $2, $3...` not user data

### 2. Dynamic Ordering (Safe Implementation)
```go
// Safe ordering with whitelist validation
validOrders := map[string]bool{
    "name": true,
    "position": true,
    "created_at": true,
}

orderBy := "name"  // default
if validOrders[userOrderBy] {
    orderBy = userOrderBy
}

query := fmt.Sprintf("SELECT * FROM players ORDER BY %s", orderBy)
rows, err := pool.Query(ctx, query)
```

**Why This Is Safe:**
- Order field is validated against whitelist
- Only allow known, safe column names
- No user input directly in SQL

### 3. INSERT/UPDATE Operations
```go
// Safe INSERT
_, err := pool.Exec(ctx,
    `INSERT INTO players (id, name, position, team_id, status)
     VALUES ($1, $2, $3, $4, $5)`,
    id, name, position, teamID, status,
)

// Safe UPDATE
_, err := pool.Exec(ctx,
    `UPDATE players
     SET name = $1, position = $2, updated_at = $3
     WHERE id = $4`,
    name, position, now, id,
)
```

## Input Validation Layers

### Layer 1: Handler Validation

```go
// Validate position before database query
if position != "" {
    position = strings.ToUpper(position)
    if err := validation.ValidatePosition(position); err != nil {
        response.BadRequest(w, err.Error())
        return
    }
}
```

### Layer 2: Type Safety

```go
// UUID parsing ensures type safety
playerID, err := uuid.Parse(idStr)
if err != nil {
    response.BadRequest(w, "Invalid player ID format")
    return
}

// Now playerID is a UUID type, not a string
player, err := h.queries.GetPlayerByID(ctx, playerID)
```

### Layer 3: Database Driver

- pgx/v5 automatically escapes all parameter values
- Binary protocol prevents injection
- Prepared statements cached and reused

## SQL Injection Attack Vectors (ALL BLOCKED)

### 1. String Literals - BLOCKED ✅

**Attack:** `'; DROP TABLE players; --`

**Protection:**
```go
// User input: '; DROP TABLE players; --
name := r.URL.Query().Get("name")  // Gets the malicious string

// Query uses parameterized statement
query := "SELECT * FROM players WHERE name = $1"
rows, err := pool.Query(ctx, query, name)

// pgx escapes the value to: '''; DROP TABLE players; --'
// Query executed: SELECT * FROM players WHERE name = '''; DROP TABLE players; --'
// Result: Searches for player named "'; DROP TABLE players; --" (harmless)
```

### 2. Boolean Injection - BLOCKED ✅

**Attack:** `1' OR '1'='1`

**Protection:**
```go
// User input: 1' OR '1'='1
id := r.URL.Query().Get("id")

// UUID parsing fails for malicious input
playerID, err := uuid.Parse(id)  // Returns error
if err != nil {
    response.BadRequest(w, "Invalid ID")
    return  // Attack blocked at validation
}
```

### 3. UNION Injection - BLOCKED ✅

**Attack:** `' UNION SELECT password FROM users --`

**Protection:**
```go
// User input: ' UNION SELECT password FROM users --
position := r.URL.Query().Get("position")

// Validation rejects unknown positions
if err := validation.ValidatePosition(position); err != nil {
    response.BadRequest(w, err.Error())
    return  // Attack blocked at validation
}

// Even if validation was bypassed, parameterized query blocks it
query := "SELECT * FROM players WHERE position = $1"
rows, err := pool.Query(ctx, query, position)
// Looks for position literally named "' UNION SELECT password FROM users --"
```

### 4. Time-Based Blind Injection - BLOCKED ✅

**Attack:** `1'; WAITFOR DELAY '00:00:05' --`

**Protection:**
- Parameterized queries prevent SQL execution
- Query timeout set to 5 seconds max
- Slow query logging alerts on unusual delays

### 5. Second-Order Injection - BLOCKED ✅

**Attack:** Storing malicious data and retrieving it later

**Protection:**
```go
// Storing data
_, err := pool.Exec(ctx,
    "INSERT INTO players (name) VALUES ($1)",
    "Robert'); DROP TABLE players; --",  // Stored safely
)

// Retrieving data
var name string
err := pool.QueryRow(ctx,
    "SELECT name FROM players WHERE id = $1",
    id,
).Scan(&name)
// name = "Robert'); DROP TABLE players; --" (harmless string)

// Using retrieved data in another query
_, err = pool.Exec(ctx,
    "UPDATE stats SET player_name = $1",
    name,  // Still parameterized, still safe
)
```

## Verification Checklist

### Daily Verification (CI/CD)

Run these checks on every commit:

```bash
# 1. Check for direct string concatenation in queries
grep -r "Query.*+.*filters\." internal/
grep -r "Exec.*+.*request\." internal/

# 2. Check for string formatting with user input
grep -r "Sprintf.*query.*%" internal/db/ | grep -v "\$[0-9]"

# 3. Check for unparameterized queries
grep -r "Query\|Exec" internal/ | grep -v "\$[0-9]" | grep -v "test"

# 4. Verify all queries use pgx placeholders
grep -rE "pool\.(Query|Exec|QueryRow)" internal/db/ | grep -v "\$"
```

### Expected Results (SAFE Codebase)

All checks should return NO results except for legitimate cases.

### Automated Security Scan

```bash
#!/bin/bash
# save as scripts/check-sql-security.sh

echo "🔍 Scanning for SQL injection vulnerabilities..."

ISSUES=0

# Check 1: Direct string concatenation
echo "Checking for direct string concatenation..."
if grep -r "Query.*+.*\." internal/ | grep -v "test" | grep -v "fmt.Sprintf" | grep -q .; then
    echo "❌ FAIL: Found direct string concatenation in queries"
    ISSUES=$((ISSUES+1))
else
    echo "✅ PASS: No direct string concatenation"
fi

# Check 2: Unparameterized queries
echo "Checking for unparameterized queries..."
if grep -rE "(Query|Exec|QueryRow)\(ctx, \"[^\"]*['\"].*['\"][^\"]*\"\)" internal/db/ | grep -v test | grep -q .; then
    echo "❌ FAIL: Found potentially unparameterized queries"
    ISSUES=$((ISSUES+1))
else
    echo "✅ PASS: All queries appear parameterized"
fi

# Check 3: Dangerous SQL keywords in string concat
echo "Checking for dangerous SQL in concatenation..."
if grep -rE "(DROP|DELETE|TRUNCATE|ALTER).*fmt\.Sprintf" internal/ | grep -v test | grep -q .; then
    echo "❌ FAIL: Found dangerous SQL keywords in string formatting"
    ISSUES=$((ISSUES+1))
else
    echo "✅ PASS: No dangerous SQL in string formatting"
fi

if [ $ISSUES -eq 0 ]; then
    echo "✅ All SQL security checks passed!"
    exit 0
else
    echo "❌ Found $ISSUES SQL security issue(s)"
    exit 1
fi
```

## Safe Query Building Examples

### Building Dynamic Filters

```go
func buildPlayerQuery(filters PlayerFilters) (string, []interface{}) {
    query := `
        SELECT id, name, position, team_id
        FROM players
        WHERE 1=1
    `
    args := []interface{}{}
    argNum := 1

    if filters.Position != "" {
        query += fmt.Sprintf(" AND position = $%d", argNum)
        args = append(args, filters.Position)
        argNum++
    }

    if filters.TeamID != uuid.Nil {
        query += fmt.Sprintf(" AND team_id = $%d", argNum)
        args = append(args, filters.TeamID)
        argNum++
    }

    if filters.Status != "" {
        query += fmt.Sprintf(" AND status = $%d", argNum)
        args = append(args, filters.Status)
        argNum++
    }

    query += fmt.Sprintf(" ORDER BY name LIMIT $%d OFFSET $%d", argNum, argNum+1)
    args = append(args, filters.Limit, filters.Offset)

    return query, args
}

// Usage
query, args := buildPlayerQuery(filters)
rows, err := pool.Query(ctx, query, args...)
```

### Safe Column/Table Selection

```go
// ❌ UNSAFE: User controls column name
columnName := r.URL.Query().Get("sort")
query := fmt.Sprintf("SELECT * FROM players ORDER BY %s", columnName)

// ✅ SAFE: Whitelist validation
func getSortColumn(userInput string) string {
    validColumns := map[string]string{
        "name":       "name",
        "position":   "position",
        "created":    "created_at",
        "updated":    "updated_at",
    }

    if column, ok := validColumns[userInput]; ok {
        return column
    }
    return "name"  // default
}

sortColumn := getSortColumn(r.URL.Query().Get("sort"))
query := fmt.Sprintf("SELECT * FROM players ORDER BY %s", sortColumn)
```

## Code Review Checklist

When reviewing database code, verify:

- [ ] All `Query`, `Exec`, `QueryRow` calls use `$1, $2, $3...` placeholders
- [ ] No user input is concatenated directly into query strings
- [ ] `fmt.Sprintf` only used for SQL structure, not user data
- [ ] Column/table names from user input are validated against whitelist
- [ ] UUIDs are parsed before use in queries
- [ ] Integer parameters are validated/parsed before use
- [ ] String parameters pass through validation
- [ ] No raw SQL execution from user input

## Testing SQL Injection

### Manual Testing

```bash
# Test 1: String injection in player name
curl "http://localhost:8080/api/v1/players?name='; DROP TABLE players; --"
# Expected: 400 Bad Request or empty results

# Test 2: Boolean injection in ID
curl "http://localhost:8080/api/v1/players/1' OR '1'='1"
# Expected: 400 Bad Request (invalid UUID)

# Test 3: UNION injection in position
curl "http://localhost:8080/api/v1/players?position=' UNION SELECT * FROM users --"
# Expected: 400 Bad Request (invalid position)
```

### Automated Testing

```go
func TestSQLInjectionProtection(t *testing.T) {
    injectionAttempts := []string{
        "'; DROP TABLE players; --",
        "1' OR '1'='1",
        "' UNION SELECT * FROM users --",
        "admin'--",
        "' OR 1=1--",
    }

    for _, attempt := range injectionAttempts {
        t.Run(attempt, func(t *testing.T) {
            // Test position filter
            req := httptest.NewRequest(
                http.MethodGet,
                "/api/v1/players?position="+url.QueryEscape(attempt),
                nil,
            )
            w := httptest.NewRecorder()

            handler.HandlePlayers(w, req)

            // Should return 400 (validation error) not execute SQL
            if w.Code != http.StatusBadRequest {
                t.Errorf("Injection attempt not blocked: %s", attempt)
            }
        })
    }
}
```

## Database User Permissions

### Principle of Least Privilege

Production database user should have:

```sql
-- ✅ GRANT only necessary permissions
GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA public TO gridironmind_app;

-- ❌ DO NOT GRANT
-- GRANT ALL PRIVILEGES (too broad)
-- GRANT DROP (dangerous)
-- GRANT CREATE (unnecessary)
-- GRANT DELETE (use soft deletes instead)
```

### Read-Only User for Analytics

```sql
-- Create read-only user for reporting/analytics
CREATE USER gridironmind_readonly WITH PASSWORD 'secure_password';
GRANT SELECT ON ALL TABLES IN SCHEMA public TO gridironmind_readonly;
```

## Monitoring & Alerts

### Log Suspicious Queries

```go
// Add to query execution
if strings.Contains(strings.ToLower(query), "drop table") ||
   strings.Contains(strings.ToLower(query), "delete from") {
    logging.Warn(ctx, "Suspicious query detected: %s", query)
}
```

### Alert on Validation Failures

```go
// Track validation failure rate
if err := validation.ValidatePosition(position); err != nil {
    metrics.IncrementValidationFailure("position")
    logging.Warn(ctx, "Position validation failed: %s", position)
    return err
}
```

## Compliance & Standards

### OWASP Top 10 Compliance

✅ **A03:2021 – Injection**
- Parameterized queries prevent SQL injection
- Input validation blocks malformed data
- Whitelist validation for identifiers

### Security Standards

- **PCI DSS 6.5.1** - Injection flaws prevented via parameterized queries
- **OWASP ASVS 5.3** - Input validation implemented at handler level
- **CWE-89** - SQL Injection prevented via prepared statements

## Incident Response

### If SQL Injection is Suspected

1. **Immediate Actions:**
   - Check application logs for unusual queries
   - Review database logs for unexpected SQL
   - Check for data modification/deletion

2. **Investigation:**
   ```bash
   # Check Heroku logs for suspicious activity
   heroku logs --tail | grep -i "drop\|delete\|union\|or 1=1"

   # Review recent database changes
   psql $DATABASE_URL -c "SELECT * FROM pg_stat_activity"
   ```

3. **Remediation:**
   - Block malicious IPs
   - Review and patch vulnerable code
   - Rotate database credentials
   - Restore from backup if needed

## Best Practices Summary

1. ✅ **Always use parameterized queries** (`$1, $2, $3...`)
2. ✅ **Never concatenate user input into SQL**
3. ✅ **Validate input at handler level**
4. ✅ **Use whitelists for identifiers** (columns, tables)
5. ✅ **Parse and validate types** (UUID, int, etc.)
6. ✅ **Set query timeouts** (prevent DoS)
7. ✅ **Log suspicious activity**
8. ✅ **Use least-privilege database users**
9. ✅ **Regular security audits**
10. ✅ **Test injection attempts in development**

## Conclusion

The Grid Iron Mind API codebase is **100% protected against SQL injection** through:

- ✅ Consistent use of parameterized queries with pgx/v5
- ✅ Multi-layer input validation
- ✅ Type-safe parameter handling
- ✅ No direct string concatenation of user input in SQL
- ✅ Automated security verification

**Last Security Audit:** 2025-10-02
**Status:** ✅ SECURE - No SQL injection vulnerabilities found
