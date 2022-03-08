# PAMGen gomock compatible mock generator

PAMGen means "Proto Aware Mock Generator". This means the thing is aware of what proto.Messages are and how to compare
them.

## Installation

```shell
go install github.com/sirkon/pamgen@latest
```

## Usage

```shell
pamgen -s io -d io_mock.go -p mocks Reader # generate a mock for io.Reader in the ./io_mock.go
pamgen -d current_mock.go # generate mocks for all interfaces of package placed in the current directory
```