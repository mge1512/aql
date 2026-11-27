# pcd-spec-sha256: 2fafcf9ca530438a6fb11f674789ae149aa9e4e48961b854a40a93e0c0cf6a6f
#
# spec file for package libaql
#
# The Go binding of libaql is a Go source library. The package installs
# the sources of the module for the build of the servers that use it.

%define goipath github.com/mge1512/aql/go/aql
%define gosrcdir %{_datadir}/go/src/%{goipath}

Name:           libaql
Version:        0.6.0
Release:        0
Summary:        Front end of the query language AQL for MCP servers
License:        MIT
URL:            https://github.com/mge1512/aql
Source0:        %{name}-%{version}.tar.gz
BuildRequires:  go >= 1.24
BuildRequires:  lemongrass = 0.1.1
BuildRequires:  re2c = 4.6
BuildRequires:  make

%description
libaql reads the text of one AQL statement for an MCP server. AQL is a
small SQL dialect: SELECT and DESCRIBE read, INSERT, UPDATE and DELETE
change one record. The library returns the statement, or an error whose
message says what to write instead.

%package devel
Summary:        Go sources of libaql
BuildArch:      noarch
Provides:       golang(%{goipath}) = %{version}

%description devel
The Go module %{goipath} of libaql, for the build of servers that use
it. The package has the sources only; there is no executable.

%prep
%autosetup -n %{name}-%{version}

%build
cd go/aql
make build LEMONGRASS_TEMPLATE=%{_datadir}/lemongrass/lempar.go

%check
cd go/aql
make test LEMONGRASS_TEMPLATE=%{_datadir}/lemongrass/lempar.go

%install
install -d %{buildroot}%{gosrcdir}
cd go/aql
install -m 0644 go.mod VERSION *.go %{buildroot}%{gosrcdir}/

%files devel
%license go/aql/LICENSE
%doc go/aql/README.md
%dir %{_datadir}/go
%dir %{_datadir}/go/src
%{_datadir}/go/src/github.com

%changelog
