clean:
	git clean -xfd
	cd ./frontend && git clean -xfd

init:
	#go mod tidy -v
	git submodule init
	git submodule update --recursive
	cd frontend && npm ci 
