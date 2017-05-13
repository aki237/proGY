%define debug_package %{nil}
Name:           proGY
Version:        2.0.3
Release:        1%{?dist}
Summary:        Simple proxy authenticator

License:        GPLv3+
URL:            https://github.com/aki237/proGY
Source0:        https://github.com/aki237/proGY/archive/2.0.3.tar.gz
Prefix:         %{_prefix}
Packager:       Akilan Elango
#BuildRequires:  golang

%description
A simple proxy authenticator for proxy connections with basic
authentication.

%prep
%autosetup


%build
CGO_ENABLED=0 go build -o proGY -ldflags "-s -w"

%install
echo "cat /usr/lib/systemd/system/proGY.service.in | sed 's/\[USER\]/\$USER/g' > /usr/lib/systemd/system/proGY.service" > proGY-setup-user.sh
install -Dm755 proGY-setup-user.sh $RPM_BUILD_ROOT/usr/bin/proGY-setup-user.sh
install -Dm755 proGY $RPM_BUILD_ROOT/usr/bin/proGY
install -Dm644 proGY.service $RPM_BUILD_ROOT/usr/lib/systemd/system/proGY.service.in


%files
/usr/bin/proGY-setup-user.sh
/usr/bin/proGY
/usr/lib/systemd/system/proGY.service.in

%post
echo "Make sure that you run 'proGY-setup-user.sh' command as the normal user as running proGY as root is not advised."

%changelog
* Sat May 13 2017 Akilan Elango <akilan1997@gmail.com>
- Added control unix domain socket
