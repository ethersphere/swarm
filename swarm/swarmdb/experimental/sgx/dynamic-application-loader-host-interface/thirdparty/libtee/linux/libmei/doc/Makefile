SED := sed
DOXYGEN := doxygen

MANUALS := html/index.html
all: $(MANUALS)

Doxyfile: Doxyfile.in
	$(SED) "s;@srcdir@;$(SRCDIR);" $^ > $@

html/index.html: Doxyfile
	$(DOXYGEN)

clean:
	$(RM) Doxyfile
	$(RM) -r html

.PHONY: clean all
