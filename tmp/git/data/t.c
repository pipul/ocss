#include <stdio.h>

struct page {
	struct {
		union {
			long a;
			int b;
		};
		int c;
	};
	int d;
};


int main() {
	struct page p;
	p.a = 10;
	p.c = 20;
	p.d = 30;
	printf("%d %d %d\n", p.a, p.c, p.d);
}
haha
hahahahahahaha
hahahahahahaha
haha
hahaweiwei
