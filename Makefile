TEST?=$$(go list ./... | grep -v 'vendor')
HOST=github.com
NAME=crunchybridge
BINARY=terraform-provider-crunchybridge
TAG_VER?=0.1.0
RELTEMPDIR?=release_tmp
CMD_SHA256SUM?="sha256sum" # sometimes 'shasum -a 256'

# Define for release to work
# GPG_USER_ID
# GPG_PASSPHRASE

# Used for local development only (i.e. install target)
NAMESPACE=CrunchyData
OS_ARCH?=linux_amd64

.PHONY: default build install release test testacc clean

default: build

build: version.txt
	go build -o ${BINARY}

clean:
	rm -rf ${RELTEMPDIR}

version.txt:
	./tools/autoversion.sh

install: build
	mkdir -p ~/.terraform.d/plugins/${HOST}/${NAMESPACE}/${NAME}/${TAG_VER}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOST}/${NAMESPACE}/${NAME}/${TAG_VER}/${OS_ARCH}

arch = darwin_amd64 darwin_arm64 linux_amd64 linux_arm64 linux_arm windows_amd64

release: version.txt $(arch:%=${BINARY}_${TAG_VER}_%.zip)
	test -n "$(GPG_USER_ID)" # $GPG_USER_ID
	test -n "$(GPG_PASSPHRASE)" # $GPG_PASSPHRASE
	cp terraform-registry-manifest.json $(RELTEMPDIR)/${BINARY}_${TAG_VER}_manifest.json
	cd $(RELTEMPDIR) && sha256sum *.zip *.json > ${BINARY}_${TAG_VER}_SHA256SUMS
	@cd $(RELTEMPDIR) && gpg --detach-sign --pinentry-mode loopback --passphrase "${GPG_PASSPHRASE}" -u "${GPG_USER_ID}" ${BINARY}_${TAG_VER}_SHA256SUMS

tokens = $(subst _, ,$*)
binary = $(word 1, $(tokens))
tgt_ver = $(word 2, $(tokens))
tgt_os = $(word 3, $(tokens))
tgt_arch = $(word 4, $(tokens))
out_binary = $(binary)_v$(tgt_ver)

%.zip: version.txt main.go internal go.mod go.sum | $(RELTEMPDIR)
	GOOS=$(tgt_os) GOARCH=$(tgt_arch) go build -o $(RELTEMPDIR)/$(out_binary)
	@cd $(RELTEMPDIR) && zip $@ $(out_binary)
	rm $(RELTEMPDIR)/$(out_binary)

$(RELTEMPDIR):
	@mkdir -pv $@

test: 
	go test -i $(TEST) || exit 1                                                   
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4                    

testacc: 
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m   
