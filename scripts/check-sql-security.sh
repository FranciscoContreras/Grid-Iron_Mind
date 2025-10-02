#!/bin/bash

# SQL Security Verification Script
# Scans codebase for SQL injection vulnerabilities

set -e

echo "🔍 Grid Iron Mind - SQL Security Scanner"
echo "========================================"
echo ""

ISSUES=0
WARNINGS=0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check 1: Direct string concatenation in queries
echo "📋 Check 1: Direct string concatenation in SQL queries..."
if grep -r "Query.*+.*\." internal/ 2>/dev/null | grep -v "test" | grep -v "fmt.Sprintf" | grep -v "//" | grep -q .; then
    echo -e "${RED}❌ FAIL${NC}: Found direct string concatenation in queries"
    grep -rn "Query.*+.*\." internal/ | grep -v "test" | grep -v "fmt.Sprintf" | grep -v "//"
    ISSUES=$((ISSUES+1))
else
    echo -e "${GREEN}✅ PASS${NC}: No direct string concatenation in queries"
fi
echo ""

# Check 2: Unparameterized queries (queries without $ placeholders)
echo "📋 Check 2: Parameterized query verification..."
UNPARAMETERIZED=$(grep -rE "\.Query\(ctx, \"[^\"]*\"[^$]*\)" internal/db/ 2>/dev/null | grep -v "test" | grep -v "\$" | wc -l | tr -d ' ')
if [ "$UNPARAMETERIZED" -gt 0 ]; then
    echo -e "${RED}❌ FAIL${NC}: Found $UNPARAMETERIZED potentially unparameterized queries"
    grep -rnE "\.Query\(ctx, \"[^\"]*\"[^$]*\)" internal/db/ | grep -v "test" | grep -v "\$"
    ISSUES=$((ISSUES+1))
else
    echo -e "${GREEN}✅ PASS${NC}: All queries use parameterized statements"
fi
echo ""

# Check 3: Dangerous SQL keywords in string formatting
echo "📋 Check 3: Dangerous SQL in string concatenation..."
if grep -rE "(DROP|DELETE|TRUNCATE|ALTER).*fmt\.Sprintf" internal/ 2>/dev/null | grep -v "test" | grep -v "//" | grep -q .; then
    echo -e "${RED}❌ FAIL${NC}: Found dangerous SQL keywords in string formatting"
    grep -rnE "(DROP|DELETE|TRUNCATE|ALTER).*fmt\.Sprintf" internal/ | grep -v "test" | grep -v "//"
    ISSUES=$((ISSUES+1))
else
    echo -e "${GREEN}✅ PASS${NC}: No dangerous SQL in string formatting"
fi
echo ""

# Check 4: Verify pgx usage (should use pool.Query, pool.Exec, pool.QueryRow)
echo "📋 Check 4: pgx driver usage verification..."
TOTAL_QUERIES=$(grep -rE "(Query|Exec|QueryRow)" internal/db/ 2>/dev/null | grep -v "test" | wc -l | tr -d ' ')
PGX_QUERIES=$(grep -rE "pool\.(Query|Exec|QueryRow)" internal/db/ 2>/dev/null | grep -v "test" | wc -l | tr -d ' ')

if [ "$TOTAL_QUERIES" -eq "$PGX_QUERIES" ]; then
    echo -e "${GREEN}✅ PASS${NC}: All $TOTAL_QUERIES queries use pgx driver"
else
    echo -e "${YELLOW}⚠️  WARN${NC}: Found non-pgx database calls"
    WARNINGS=$((WARNINGS+1))
fi
echo ""

# Check 5: Verify UUID parsing before database queries
echo "📋 Check 5: UUID validation before queries..."
UUID_PARSE_COUNT=$(grep -r "uuid.Parse" internal/handlers/ 2>/dev/null | wc -l | tr -d ' ')
if [ "$UUID_PARSE_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✅ PASS${NC}: Found $UUID_PARSE_COUNT UUID validations in handlers"
else
    echo -e "${YELLOW}⚠️  WARN${NC}: No UUID parsing found in handlers"
    WARNINGS=$((WARNINGS+1))
fi
echo ""

# Check 6: Input validation usage
echo "📋 Check 6: Input validation checks..."
VALIDATION_COUNT=$(grep -r "validation\." internal/handlers/ 2>/dev/null | grep -v "import" | wc -l | tr -d ' ')
if [ "$VALIDATION_COUNT" -gt 10 ]; then
    echo -e "${GREEN}✅ PASS${NC}: Found $VALIDATION_COUNT validation calls"
else
    echo -e "${YELLOW}⚠️  WARN${NC}: Low validation usage: $VALIDATION_COUNT calls"
    WARNINGS=$((WARNINGS+1))
fi
echo ""

# Check 7: Verify no raw SQL execution
echo "📋 Check 7: Raw SQL execution check..."
if grep -rE "db\.Exec\(.*\".*SELECT|INSERT|UPDATE|DELETE" internal/ 2>/dev/null | grep -v "\$" | grep -v "test" | grep -q .; then
    echo -e "${RED}❌ FAIL${NC}: Found raw SQL execution without parameters"
    ISSUES=$((ISSUES+1))
else
    echo -e "${GREEN}✅ PASS${NC}: No raw SQL execution found"
fi
echo ""

# Check 8: Verify PreparedStatement usage pattern
echo "📋 Check 8: Query parameter placeholders..."
PLACEHOLDER_COUNT=$(grep -rE "\\\$[0-9]+" internal/db/ 2>/dev/null | wc -l | tr -d ' ')
if [ "$PLACEHOLDER_COUNT" -gt 50 ]; then
    echo -e "${GREEN}✅ PASS${NC}: Found $PLACEHOLDER_COUNT parameterized query placeholders"
else
    echo -e "${YELLOW}⚠️  WARN${NC}: Low placeholder usage: $PLACEHOLDER_COUNT"
    WARNINGS=$((WARNINGS+1))
fi
echo ""

# Check 9: No direct user input in query strings
echo "📋 Check 9: User input in SQL queries..."
if grep -rE "Query.*\+.*filters\.|Query.*\+.*request\.|Query.*\+.*params\." internal/db/ 2>/dev/null | grep -v "test" | grep -q .; then
    echo -e "${RED}❌ FAIL${NC}: Found user input concatenated in queries"
    grep -rnE "Query.*\+.*filters\.|Query.*\+.*request\.|Query.*\+.*params\." internal/db/ | grep -v "test"
    ISSUES=$((ISSUES+1))
else
    echo -e "${GREEN}✅ PASS${NC}: No user input concatenated in queries"
fi
echo ""

# Check 10: Verify context usage in queries
echo "📋 Check 10: Context usage in database queries..."
CONTEXT_QUERIES=$(grep -rE "\.(Query|Exec|QueryRow)\(ctx," internal/db/ 2>/dev/null | wc -l | tr -d ' ')
if [ "$CONTEXT_QUERIES" -gt 20 ]; then
    echo -e "${GREEN}✅ PASS${NC}: Found $CONTEXT_QUERIES queries with context"
else
    echo -e "${YELLOW}⚠️  WARN${NC}: Low context usage in queries: $CONTEXT_QUERIES"
    WARNINGS=$((WARNINGS+1))
fi
echo ""

# Summary
echo "========================================"
echo "📊 Security Scan Summary"
echo "========================================"
echo ""

if [ $ISSUES -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo -e "${GREEN}✅ EXCELLENT${NC}: All SQL security checks passed!"
    echo "   - No SQL injection vulnerabilities found"
    echo "   - All queries use parameterized statements"
    echo "   - Input validation in place"
    echo ""
    exit 0
elif [ $ISSUES -eq 0 ]; then
    echo -e "${YELLOW}⚠️  GOOD${NC}: No critical issues found"
    echo "   - Warnings: $WARNINGS"
    echo "   - Consider addressing warnings for best practices"
    echo ""
    exit 0
else
    echo -e "${RED}❌ FAILED${NC}: SQL security issues detected"
    echo "   - Critical Issues: $ISSUES"
    echo "   - Warnings: $WARNINGS"
    echo ""
    echo "Please review and fix the issues above before deploying."
    exit 1
fi
