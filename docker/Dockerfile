FROM ubuntu:16.04

MAINTAINER Takuya Kuwahara <taakuu19@gmail.com>

ENV PATH   /usr/local/go/bin:$PATH
ENV GOPATH /go

# Make sure apt-get is up to date and dependent packages are installed
RUN \
  apt-get update && \
  apt-get install -y automake autogen build-essential ca-certificates           \
    gcc-5-multilib g++-5-multilib clang llvm-dev                                \
    libtool libxml2-dev uuid-dev libssl-dev swig openjdk-8-jdk pkg-config patch \
    make xz-utils cpio wget zip unzip p7zip git mercurial bzr texinfo help2man  \
    libgcrypt11-dev cmake libaio-dev libncurses5-dev libev-dev                  \
    libcurl4-gnutls-dev vim-common                                              \
    --no-install-recommends

# Configure the container for OSX cross compilation
ENV OSX_SDK     MacOSX10.11.sdk
ENV OSX_NDK_X86 /usr/local/osx-ndk-x86

RUN \
  OSX_SDK_PATH=https://s3.dockerproject.org/darwin/v2/$OSX_SDK.tar.xz && \
  wget -q $OSX_SDK_PATH && \
  git clone https://github.com/tpoechtrager/osxcross.git && \
  mv `basename $OSX_SDK_PATH` /osxcross/tarballs/        && \
  \
  sed -i -e 's|-march=native||g' /osxcross/build_clang.sh /osxcross/wrapper/build.sh && \
  UNATTENDED=yes OSX_VERSION_MIN=10.6 /osxcross/build.sh                             && \
  mv /osxcross/target $OSX_NDK_X86                                                   && \
  \
  rm -rf /osxcross

ENV PATH $OSX_NDK_X86/bin:$PATH

# Install go 1.8.1
RUN \
  wget -q https://storage.googleapis.com/golang/go1.8.1.linux-amd64.tar.gz && \
  tar -C /usr/local -xzf `basename https://storage.googleapis.com/golang/go1.8.1.linux-amd64.tar.gz` && \
  rm -f `basename $ROOT_DIST`
ENV GOROOT /usr/local/go
