all:
	echo haha >> env.sh
	echo haha >> test/demo.go
	#echo haha >> README.md
	#echo haha >> tmp/git/test.txt
	git add .
	git commit -m "update"
	git push origin feature/b_fd_2

