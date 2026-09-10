target:
	test -f x != y
	X=1 ./cmd
	echo a!=b
	[ ! -d dir ] || rm -r dir
