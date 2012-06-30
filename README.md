`DiplomaEnhancer` is a daemon which manipulates your local host file to block webpages which you consider distracting.

# Compiling
Because of the static content inside the repo, a mere `go get` won't work (it will compile, but the admin interface will not display correctly). Use

	git clone git@github.com:surma/diplomaenhancer
	cd diplomaenhancer
	go get .

instead.

	sudo ./diplomaenhancer

should start the daemon right up.
