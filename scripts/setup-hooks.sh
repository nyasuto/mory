#!/bin/bash
# Git hooks setup script for Mory project
# Sets up pre-commit hook to enforce branch strategy and run quality checks

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔧 Setting up Git hooks for Mory project...${NC}"

# Get project root directory
PROJECT_ROOT=$(git rev-parse --show-toplevel 2>/dev/null) || {
    echo -e "${RED}❌ Error: Not in a Git repository${NC}"
    exit 1
}

# Define hook paths
HOOKS_DIR="$PROJECT_ROOT/.git/hooks"
SOURCE_HOOK="$PROJECT_ROOT/scripts/pre-commit"
TARGET_HOOK="$HOOKS_DIR/pre-commit"

# Check if source hook exists
if [[ ! -f "$SOURCE_HOOK" ]]; then
    echo -e "${RED}❌ Error: Source pre-commit hook not found at $SOURCE_HOOK${NC}"
    exit 1
fi

# Create hooks directory if it doesn't exist
if [[ ! -d "$HOOKS_DIR" ]]; then
    echo -e "${YELLOW}⚠️  Git hooks directory not found, creating...${NC}"
    mkdir -p "$HOOKS_DIR"
fi

# Backup existing pre-commit hook if it exists
if [[ -f "$TARGET_HOOK" ]]; then
    echo -e "${YELLOW}⚠️  Existing pre-commit hook found, backing up...${NC}"
    cp "$TARGET_HOOK" "$TARGET_HOOK.backup.$(date +%Y%m%d_%H%M%S)"
    echo -e "${GREEN}✅ Backup created: $TARGET_HOOK.backup.$(date +%Y%m%d_%H%M%S)${NC}"
fi

# Copy and set permissions
echo -e "${BLUE}📁 Installing pre-commit hook...${NC}"
cp "$SOURCE_HOOK" "$TARGET_HOOK"
chmod +x "$TARGET_HOOK"

# Verify installation
if [[ -x "$TARGET_HOOK" ]]; then
    echo -e "${GREEN}✅ Pre-commit hook successfully installed!${NC}"
else
    echo -e "${RED}❌ Error: Failed to install pre-commit hook${NC}"
    exit 1
fi

# Test the hook
echo -e "${BLUE}🧪 Testing hook functionality...${NC}"

# Create a temporary test
TEST_BRANCH=$(git symbolic-ref --short HEAD 2>/dev/null || echo "detached")
echo -e "${BLUE}Current branch: $TEST_BRANCH${NC}"

if [[ "$TEST_BRANCH" == "main" ]]; then
    echo -e "${YELLOW}⚠️  You're on the main branch. The hook will prevent commits to main.${NC}"
    echo -e "${BLUE}To test the hook, switch to a feature branch:${NC}"
    echo -e "  git checkout -b test/hook-functionality"
else
    echo -e "${GREEN}✅ Branch check: Ready for testing${NC}"
fi

# Display hook features
echo -e "${BLUE}📋 Installed hook features:${NC}"
echo -e "  ✅ Prevents direct commits to main branch"
echo -e "  ✅ Validates branch naming conventions"
echo -e "  ✅ Runs 'make quality' before commit"
echo -e "  ✅ Checks for sensitive information"
echo -e "  ✅ Provides helpful error messages"

echo -e "${GREEN}🎉 Git hooks setup complete!${NC}"
echo -e "${BLUE}Usage:${NC}"
echo -e "  - Normal commits will now trigger quality checks automatically"
echo -e "  - To bypass hooks in emergencies: git commit --no-verify"
echo -e "  - To uninstall: rm $TARGET_HOOK"

echo -e "${YELLOW}💡 Tip: Run 'make quality' manually to fix issues before committing${NC}"

exit 0