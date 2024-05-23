.SILENT:

new-project:
	@cp -r ./template ./$(var)
	sed -i 's|packageName|$(var)|' ./$(var)/Makefile