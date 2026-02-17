ARTIFACT_ID=ces-exporter
MAKEFILES_VERSION=10.5.0
VERSION=1.3.0

GOTAG=1.26.0
MOCKERY_VERSION=v2.53.0
.DEFAULT_GOAL:=help

## Image URL to use all building/pushing image targets
IMAGE?=cloudogu/${ARTIFACT_ID}:${VERSION}

K8S_RESOURCE_DIR=${WORKDIR}/k8s
K8S_COMPONENT_SOURCE_VALUES = ${HELM_SOURCE_DIR}/values.yaml
K8S_COMPONENT_TARGET_VALUES = ${HELM_TARGET_DIR}/values.yaml
HELM_PRE_GENERATE_TARGETS = helm-values-update-image-version
HELM_POST_GENERATE_TARGETS = helm-values-replace-image-repo template-stage template-log-level template-image-pull-policy template-importer-public-key
CHECK_VAR_TARGETS=check-all-vars
IMAGE_IMPORT_TARGET=image-import
IMAGE=ces-exporter:${VERSION}

include build/make/variables.mk
PREPARE_PACKAGE=$(DEBIAN_CONTENT_DIR)/control/postinst $(DEBIAN_CONTENT_DIR)/control/postrm prepare-classic-docker

ADDITIONAL_CLEAN=clean_charts
clean_charts:
	rm -rf ${K8S_HELM_RESSOURCES}/charts

include build/make/dependencies-gomod.mk
include build/make/build.mk
include build/make/test-common.mk
include build/make/test-unit.mk
include build/make/static-analysis.mk
include build/make/clean.mk
include build/make/mocks.mk
include build/make/release.mk
include build/make/self-update.mk
include build/make/k8s-component.mk
include build/make/package-debian.mk
include build/make/deploy-debian.mk

$(DEBIAN_CONTENT_DIR)/control/postinst: $(DEBIAN_CONTENT_DIR)/control
	@install -p -m 0755 $(WORKDIR)/deb/DEBIAN/postinst $@

$(DEBIAN_CONTENT_DIR)/control/postrm: $(DEBIAN_CONTENT_DIR)/control
	@install -p -m 0755 $(WORKDIR)/deb/DEBIAN/postrm $@

.PHONY: prepare-classic-docker
prepare-classic-docker:
	echo "Classic docker build for ces-exporter:${VERSION}"

# delete binary, if it exists, to avoid adding it to the debian-package
	@if [ -f $(BINARY) ]; then \
		echo "Removing binary from $(BINARY)"; \
		rm -f $(BINARY); \
	fi

	mkdir -p ${DEBIAN_CONTENT_DIR}/data/tmp/
	rm -f ${DEBIAN_CONTENT_DIR}/data/tmp/exporter-image.tar
	docker build -t ${IMAGE} --target classic .
	docker image save -o ${DEBIAN_CONTENT_DIR}/data/tmp/exporter-image.tar ${IMAGE}

.PHONY: mocks
mocks: ${MOCKERY_BIN} ${MOCKERY_YAML} ## target is used to generate mocks for all interfaces in a project.
	${MOCKERY_BIN}
	@echo "Mocks successfully created."

.PHONY: helm-values-update-image-version
helm-values-update-image-version: $(BINARY_YQ)
	@echo "Updating the image version in source values.yaml to ${VERSION}..."
	@$(BINARY_YQ) -i e ".image.tag = \"${VERSION}\"" ${K8S_COMPONENT_SOURCE_VALUES}

.PHONY: helm-values-replace-image-repo
helm-values-replace-image-repo: $(BINARY_YQ)
	@if [[ ${STAGE} == "development" ]]; then \
		echo "Setting dev image repo in target values.yaml!" ;\
		$(BINARY_YQ) -i e ".image.registry=\"$(shell echo '${IMAGE_DEV}' | sed 's/\([^\/]*\)\/\(.*\)/\1/')\"" ${K8S_COMPONENT_TARGET_VALUES} ;\
		$(BINARY_YQ) -i e ".image.repository=\"$(shell echo '${IMAGE_DEV}' | sed 's/\([^\/]*\)\/\(.*\)/\2/')\"" ${K8S_COMPONENT_TARGET_VALUES} ;\
	fi

.PHONY: template-stage
template-stage: $(BINARY_YQ)
	@if [[ ${STAGE} == "development" ]]; then \
		echo "Setting STAGE env in deployment to ${STAGE}!" ;\
		$(BINARY_YQ) -i e ".env.stage=\"${STAGE}\"" ${K8S_COMPONENT_TARGET_VALUES} ;\
	fi

.PHONY: template-log-level
template-log-level: ${BINARY_YQ}
	@if [[ "${STAGE}" == "development" ]]; then \
	  echo "Setting LOG_LEVEL env in deployment to ${LOG_LEVEL}!" ; \
	  $(BINARY_YQ) -i e ".env.logLevel=\"${LOG_LEVEL}\"" "${K8S_COMPONENT_TARGET_VALUES}" ; \
	fi

.PHONY: template-image-pull-policy
template-image-pull-policy: $(BINARY_YQ)
	@if [[ "${STAGE}" == "development" ]]; then \
		  echo "Setting pull policy to always!" ; \
		  $(BINARY_YQ) -i e ".imagePullPolicy=\"Always\"" "${K8S_COMPONENT_TARGET_VALUES}" ; \
	fi

.PHONY: template-importer-public-key
template-importer-public-key: $(BINARY_YQ)
	@if [[ "${STAGE}" == "development" ]]; then \
		  echo "Setting importer-public-key from environment-variable 'IMPORTER_PUBLIC_KEY'" ; \
		  $(BINARY_YQ) -i e ".publicKey.data=\"${IMPORTER_PUBLIC_KEY}\"" "${K8S_COMPONENT_TARGET_VALUES}" ; \
	fi

.PHONY: apikey-secret
apikey-secret: $(BINARY_YQ) ## generates a K8s secret for the API key from an environment variable
	@kubectl create secret generic ces-exporter-api --from-literal=apiKey=${EXPORTER_API_KEY} --namespace="${NAMESPACE}" --context="${KUBE_CONTEXT_NAME}"

.PHONY: helm-apply-dev
helm-apply-dev:
	@sed -i -E "s/(^VERSION=[[:digit:]].[[:digit:]].[[:digit:]])/\1-$$(date +%s)/g" Makefile
	@make helm-apply
	@sed -i -E "s/(^VERSION=[[:digit:]].[[:digit:]].[[:digit:]])-.*/\1/g" Makefile
	@sed -i -E "s/(tag: [[:digit:]].[[:digit:]].[[:digit:]])-.*/\1/g" k8s/helm/values.yaml