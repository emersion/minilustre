CLANG ?= clang
TARGETS = pendulum simple sujet tutorial
CFLAGS ?= -Wall -Wextra -Wno-unused-parameter

all: $(TARGETS)

%.ll: %.mls
	go run ../cmd/minilustre <$^ >$@

%.o: %.ll
	$(CLANG) $(CFLAGS) -c -o $@ $^

stdlib.o: stdlib.c
	$(CC) $(CFLAGS) -c -o $@ $^

%: %.o %-main.o stdlib.o
	$(CC) $(CFLAGS) -o $@ $^

.PHONY: clean
clean:
	$(RM) -f $(TARGETS) *.o *.ll
