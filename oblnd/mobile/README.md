## Building mobile libraries

### Prerequisites
#### protoc
Download the executable binary from  https://github.com/protocolbuffers/protobuf/releases for the platform.  Make it in your system environment variable `PATH`.
Any `protoc` version 3.0+ is ok.

```bash
protoc --version
  libprotoc 3.12.4
go install  google.golang.org/protobuf/cmd/protoc-gen-go  google.golang.org/grpc/cmd/protoc-gen-go-grpc
# protoc-gen-go and  protoc-gen-go-grpc are used with protoc to generate grpc code from proto file.
```

#### android sdk ndk
SDK(NDK) can be installed via Android Studio. If you are not using Android Studio, to work with bindings for Android, you need to have [Android SDK](https://developer.android.com/sdk/index.html#Other) installed and the `ANDROID_HOME` environment variable set to the SDK path. You also need the [NDK](https://developer.android.com/ndk/) installed; The easiest way is to run the SDK command `sdkmanager ndk-bundle`.  


#### gomobile
```bash
go  install  golang.org/x/mobile/cmd/gomobile  

#more document as below
```
Follow the [gomobile instruction](https://github.com/golang/go/wiki/Mobile) to install the `gomobile` package and dependencies.

Remember to run `gomobile init` (otherwise the `obd` build might just hang).

Note that `gomobile` only supports building projects from `GOPATH` at this point.

#### falafel
Install [`falafel`](https://github.com/lightninglabs/falafel):
```shell
⛰  go get -u -v github.com/lightninglabs/falafel
# falafel will be used to genarate in-memory gRPC client code for mobile
```

### gen falafel's go code bind for mobile
```bash
go  get -d  golang.org/x/mobile/cmd/gomobile
cd ./mobile && ./gen_bindings.sh v0.9.1
```

### Building `obd` for iOS
ios build should install xcode and swift; more docs reference [gomobile](https://github.com/golang/go/wiki/Mobile) official docs.
```shell
⛰  make ios
```

### Building `obd` for Android
```shell
⛰  make android
# Now the `ObdMobile.aar` will be built and output under `./mobile/build/android`; `ObdMobile.aar` is the library that can be imported into an android project.
```

`make mobile` will build both iOS and Android libs.

### Libraries
After the build has succeeded, the libraries will be found in
`mobile/build/ios/Obdmobile.framework` and
`mobile/build/android/ObdMobile.aar`. Reference your platforms' SDK
documentation for how to add the library to your project.

#### Generating proto definitions for your language.
In order to call the methods in the generated library, the serialized proto for the given RPC call must be provided. Similarly, the response will be a serialized proto.

##### iOS

In order to generate protobuf definitions for iOS, add `--swift_out=.` to the first `protoc` invocation found in [`gen_protos.sh`](../lnrpc/gen_protos.sh).

Then, some changes to [Dockerfile]((../lnrpc/Dockerfile)) need to be done in order to use the [Swift protobuf](https://github.com/apple/swift-protobuf) plugin with protoc:

1. Replace the base image with `FROM swift:focal` so that Swift can be used.
2. `clang-format='1:7.0*'` is unavailable in Ubuntu Focal. Change that to
`clang-format='1:10.0*`.
3. On the next line, install Go and set the environment variables by adding the following commands:

```
RUN apt-get install -y wget \
    && wget -c https://golang.org/dl/go1.17.2.linux-amd64.tar.gz -O - \
    | tar -xz -C /usr/local
ENV GOPATH=/go
ENV PATH=$PATH:/usr/local/go/bin:/go/bin
```

4. At the end of the file, just above `CMD`, add the following `RUN` command.
This will download and compile the latest tagged release of Swift protobuf.

```
RUN git clone https://github.com/apple/swift-protobuf.git \
&& cd swift-protobuf \ 
&& git checkout $(git describe --tags --abbrev=0) \
&& swift build -c release \
&& mv .build/release/protoc-gen-swift /bin
```

Finally, run `make rpc`.

##### Android

In order to generate protobuf definitions for Android, add `--java_out=.`
to the first `protoc` invocation found in
[`gen_protos.sh`](../lnrpc/gen_protos.sh). Then, run `make rpc`.

### Options
Similar to lnd, subservers can be conditionally compiled with the build by
setting the tags argument:

```shell
⛰  make ios
```

To support subservers that have APIs with name conflicts, pass the "prefix"
flag. This will add the subserver name as a prefix to each method name:

```shell
⛰  make ios prefix=1
```

### API docs
TODO(healergyl, Ben)

The mobile API docs is at https:https://github.com/omnilaboratory/obd/tree/master/docs/mobile.md
