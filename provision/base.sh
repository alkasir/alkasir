echo install base

apt-get install -y postgresql-client
apt-get install -y --no-install-recommends \
          build-essential \
          ca-certificates \
          libgtk-3-dev \
          libtool \
          libappindicator3-dev \
          mercurial \
          patch \
          pkg-config \
          unzip \
          uuid-dev \
          wget \
          xz-utils \
          icnsutils \
          libbz2-dev \
          python-pip \
        ;
