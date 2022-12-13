FROM ubuntu:20.04

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update --fix-missing
RUN apt-get install -y \
  git \
  yasm \
  python \
  gcc \
  g++ \
  cmake \
  make \
  curl \
  wget \
  apt-transport-https \
  m4 \
  zip \
  unzip \
  vim \
  build-essential

# Install dependencies of SCALE-MAMBA
WORKDIR /
RUN mkdir -p /local
WORKDIR /local

# install MPIR
RUN curl -O 'http://mpir.org/mpir-3.0.0.tar.bz2'
RUN tar xf mpir-3.0.0.tar.bz2
WORKDIR /local/mpir-3.0.0
RUN ./configure --enable-cxx --prefix="/local/mpir"
RUN make && make check && make install

# install OpenSSL 1.1.0
WORKDIR /local
RUN curl -O https://www.openssl.org/source/old/1.1.1/openssl-1.1.1e.tar.gz
RUN tar -xf openssl-1.1.1e.tar.gz
WORKDIR /local/openssl-1.1.1e
RUN ./config --prefix="/local/openssl"
RUN make && make install

# install crypto++
WORKDIR /local
RUN echo bla
RUN curl -O https://www.cryptopp.com/cryptopp800.zip
RUN unzip cryptopp800.zip -d cryptopp800
WORKDIR /local/cryptopp800
RUN make && make install PREFIX="/local/cryptopp"

ENV PATH="/local/openssl/bin/:${PATH}"
ENV C_INCLUDE_PATH="/local/openssl/include/:${C_INCLUDE_PATH}"
ENV CPLUS_INCLUDE_PATH="/local/openssl/include/:${CPLUS_INCLUDE_PATH}"
ENV LIBRARY_PATH="/local/openssl/lib/:${LIBRARY_PATH}"
ENV LD_LIBRARY_PATH="/local/openssl/lib/:${LD_LIBRARY_PATH}"
ENV C_INCLUDE_PATH="/local/mpir/include/:${C_INCLUDE_PATH}"
ENV CPLUS_INCLUDE_PATH="/local/mpir/include/:${CPLUS_INCLUDE_PATH}"
ENV LIBRARY_PATH="/local/mpir/lib/:${LIBRARY_PATH}"
ENV LD_LIBRARY_PATH="/local/mpir/lib/:${LD_LIBRARY_PATH}"
ENV CPLUS_INCLUDE_PATH="/local/cryptopp/include/:${CPLUS_INCLUDE_PATH}"
ENV LIBRARY_PATH="/local/cryptopp/lib/:${LIBRARY_PATH}"
ENV LD_LIBRARY_PATH="/local/cryptopp/lib/:${LD_LIBRARY_PATH}"

# Get SCALE-MAMBA
WORKDIR /
RUN git clone https://github.com/KULeuven-COSIC/SCALE-MAMBA.git
WORKDIR /SCALE-MAMBA

## Configure (compile later) SCALE-MAMBA
WORKDIR /SCALE-MAMBA
RUN cp CONFIG CONFIG.mine
RUN echo 'ROOT = /SCALE-MAMBA' >> CONFIG.mine
RUN echo 'OSSL = /local/openssl' >> CONFIG.mine
RUN cp Auto-Test-Data/1/* Data/
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | bash -s -- -y
ENV PATH="/root/.cargo/bin:${PATH}"
RUN ./compile.sh Programs/tutorial/

WORKDIR /
## Install golang
RUN wget --no-check-certificate https://dl.google.com/go/go1.17.3.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.17.3.linux-amd64.tar.gz
ENV PATH="/usr/local/go/bin:${PATH}"

## Install MPCService
ADD . / /root/go/src/github.com/krakenh2020/MPCService/
WORKDIR /root/go/src/github.com/krakenh2020/MPCService
RUN go install
ENV PATH="/root/go/bin:${PATH}"
ENV GOPATH="/root/go"

# Compile Scale with new input/output and programs
RUN cp computation/scale_files/Input_Output_Simple.cpp /SCALE-MAMBA/src/Input_Output/Input_Output_Simple.cpp
RUN cp computation/scale_files/Input_Output_Simple.h /SCALE-MAMBA/src/Input_Output/Input_Output_Simple.h
RUN cp computation/scale_files/Input_Output_Base.h /SCALE-MAMBA/src/Input_Output/Input_Output_Base.h
RUN cp computation/scale_files/IO.h /SCALE-MAMBA/src/Input_Output/IO.h


RUN sed -i s/input\_share\(Proc\.read\_Ri\(r.0.\)/input_share\(Proc\.read\_Ri\(r\[0\]\),\ P.whoami\(\)/g /SCALE-MAMBA/src/Processor/Instruction.cpp
RUN #cp computation/scale_files/Instruction.cpp /SCALE-MAMBA/src/Processor/Instruction.cpp

RUN rm /SCALE-MAMBA/src/Input_Output/Input_Output_C.h
RUN rm /SCALE-MAMBA/src/Input_Output/Input_Output_C.cpp
RUN rm /SCALE-MAMBA/src/Input_Output/Input_Output_Test.cpp
RUN rm /SCALE-MAMBA/src/Input_Output/Input_Output_Test.h

RUN cp computation/scale_files/input_output.py /SCALE-MAMBA/Compiler/input_output.py
RUN cp computation/scale_files/lin_alg.py /SCALE-MAMBA/Compiler/lin_alg.py
RUN cp -r computation/scale_files/MPCService /SCALE-MAMBA/Programs
RUN cp -r computation/scale_files/Input /SCALE-MAMBA/Input
WORKDIR /SCALE-MAMBA
RUN make progs

WORKDIR /root/go/src/github.com/krakenh2020/MPCService
# Start the MPC server
ENTRYPOINT ["MPCService"]
