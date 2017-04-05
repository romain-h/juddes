NAME = juddes

build:
	@mkdir -p bin
	go build -o bin/$(NAME)

clean:
	@rm -rf bin/* pkg tmp

