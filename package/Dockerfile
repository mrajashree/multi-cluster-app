FROM ubuntu:16.04
COPY multi-cluster-app /usr/bin/
chmod +x /usr/bin/multi-cluster-app

ENV RANCHER_HELM_VERSION v2.10.0-rancher5

RUN curl -sLf https://github.com/rancher/helm/releases/download/${CATTLE_HELM_VERSION}/helm > /usr/bin/helm && \
    curl -sLf https://github.com/rancher/helm/releases/download/${CATTLE_HELM_VERSION}/tiller > /usr/bin/tiller && \
    chmod +x /usr/bin/helm /usr/bin/tiller

CMD ["multi-cluster-app"]
