.PHONY: test test-arch test-dotfiles

ARCHINSTALL_IMAGE = archinstall-validate

test-arch:
	@docker build -q -t $(ARCHINSTALL_IMAGE) -f Dockerfile.archinstall . > /dev/null
	@echo "Validando archinstall JSONs..."
	@docker run --rm -v $(PWD):/data $(ARCHINSTALL_IMAGE) /data/archinstall-manual.json && echo "OK archinstall-manual.json"
	@docker run --rm -v $(PWD):/data $(ARCHINSTALL_IMAGE) /data/archinstall-auto.json && echo "OK archinstall-auto.json"

test-dotfiles:
	docker build -t dotfiles-test -f Dockerfile.dotfiles ..
	docker run -it --rm dotfiles-test

test: test-arch test-dotfiles
