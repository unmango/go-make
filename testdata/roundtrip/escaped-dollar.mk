$$V: $$V

FOO := $$V

escaped: $$ ${FOO}

target:
	echo $$HOME
