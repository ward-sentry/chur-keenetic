FROM ubuntu:20.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
    bash \
    binutils \
    bison \
    build-essential \
    bzip2 \
    ca-certificates \
    curl \
    file \
    flex \
    gawk \
    gettext \
    git \
    libncurses-dev \
    libssl-dev \
    libz-dev \
    patch \
    perl \
    python2 \
    python3 \
    python3-distutils \
    python3-setuptools \
    rsync \
    subversion \
    time \
    unzip \
    wget \
    xz-utils \
    zlib1g-dev \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /work
