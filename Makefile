include ../make/Makefile-for-go.mk 

NAME= $(notdir $(shell pwd))
TAG=$(shell git  describe --tags --abbrev=0 )
TESTARGS=-stopAfter 20s -reportLoss

build:
	@go build -ldflags '-w -s -X main.Version=${NAME}-${TAG}' -o ~/bin/${NAME}-${TAG}
	@notify-send 'Build Complete' 'Your project has been build successfully!' -u normal -t 7500 -i checkbox-checked-symbolic

test:	${NAME}
	cat liste | xargs sudo ./${NAME} ${TESTARGS}

netup.conf:
	scp callisto.uvsq.fr:/local/xnetup/netup.conf .

netup:	netup.conf ${NAME}
	rg '^host'  netup.conf | awk '{ print $$2 }' | adnsresfilter | rg uvsq.fr | rg -v reseau | xargs sudo ./${NAME}

clean:
	@touch ${NAME} netup.conf
	@rm ${NAME} netup.conf
