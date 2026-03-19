.PHONY: test

test:
	docker build -t dotfiles-test .
	docker run -it --rm dotfiles-test
