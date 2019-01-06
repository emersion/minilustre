CLANG ?= clang
TARGETS = pendulum simple sujet tutorial

all: $(TARGETS)

%.ll: %.mls
	go run ../cmd/minilustre <$^ >$@

%.o: %.ll
	$(CLANG) -c -o $@ $^

stdlib.o: stdlib.c
	$(CC) -c -o $@ $^

%: %.o %-main.o stdlib.o
	$(CC) -o $@ $^

.PHONY: clean
clean:
	$(RM) -f $(TARGETS) *.o *.ll
