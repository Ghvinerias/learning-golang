# Go Project Management Makefile
#
# Use this Makefile to create new Go projects from a template.
# Usage: make new-project var=project-name

.SILENT:
.PHONY: new-project

# Default target
.DEFAULT_GOAL := help

# ANSI color codes for better readability
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

# Display help information
help:
	echo -e "${YELLOW}Go Project Management${NC}"
	echo -e "${GREEN}Available commands:${NC}"
	echo -e "  make new-project var=project-name  ${YELLOW}Create a new Go project${NC}"

# Create a new Go project from template
new-project:
	# Input validation
	if [ -z "$(var)" ]; then \
		echo -e "${RED}Error:${NC} Project name not specified"; \
		echo -e "Usage: make new-project var=project-name"; \
		exit 1; \
	fi
	
	# Check if project already exists
	if [ -d "./$(var)" ]; then \
		echo -e "${RED}Error:${NC} Project '$(var)' already exists"; \
		exit 1; \
	fi
	
	# Create project from template
	echo -e "${GREEN}Creating new project:${NC} $(var)"
	@cp -r ./template ./$(var)
	sed -i 's|packageName|$(var)|' ./$(var)/Makefile
	cd ./$(var) && go mod init $(var)
	echo -e "${GREEN}Project created successfully at${NC} ./$(var)"