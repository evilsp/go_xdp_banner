SUBDIRS := grafana loki prometheus
NAMESPACE ?= xdp-banner

.PHONY: all $(SUBDIRS) install uninstall include-config check-file

all: $(SUBDIRS)

$(SUBDIRS):
	@$(MAKE) install SUBDIR=$@

check-file:
	@if [ ! -f "$(SUBDIR)/config.mk" ]; then \
		echo "Error: $(SUBDIR)/config.mk not found"; \
		exit 1; \
	fi

include-config: check-file
	$(eval include $(SUBDIR)/config.mk)

install: include-config
	helm upgrade --install $(RELEASE_NAME) $(CHART_NAME) \
		--namespace $(NAMESPACE) \
		--create-namespace \
		-f $(SUBDIR)/values.yml

uninstall: include-config
	helm uninstall $(RELEASE_NAME) --namespace $(NAMESPACE)