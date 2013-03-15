all:
	echo haha >> env.sh
	echo haha >> test/demo.go
	echo haha >> README.md	
	git add .
	git commit -m "update"
	git push origin feature/b_fd_2
add:
	echo haha >> env.sh
	echo haha >> test/demo.go
	git add .
	git commit -m "update"
