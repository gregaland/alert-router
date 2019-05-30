# don't generate debug packages
%define debug_package %{nil}

Name: alert-router
Version: %{version}
Summary: Simple Alerts
Release: el7
License: Charter
Source0: %{name}-%{version}.tar.gz
Group: System/Management
BuildRoot: %{_builddir}/%{name}-root
AutoReq: no

%description
Simple alerting api

# Rig directory definitions
%define _arprefix /opt/alert-router
%define _arusr %{_prefix}/usr
%define _arvar %{_prefix}/var
%define _arbin %{_prefix}/bin
%define _arlib %{_prefix}/lib
%define _arsysconfdir %{_prefix}/etc
%define _artmpdir %{_prefix}/tmp
%define _arapps %{_prefix}/apps
%define _arlog %{_var}/log
%define _arsystemd %{_sysconfdir}/systemd/system
%define _arsysconfig %{_sysconfdir}/sysconfig
%define _arlogrotated %{_sysconfdir}/logrotate.d
%define _arrsyslogd %{_sysconfdir}/rsyslog.d

%prep

%setup

%build
make build

%install
# Create the top level directory
install -d -m 0755 %{buildroot}%{_arprefix}
install -d -m 0755 %{buildroot}%{_arusr}
install -d -m 0755 %{buildroot}%{_arvar}
install -d -m 0755 %{buildroot}%{_arbin}
install -d -m 0755 %{buildroot}%{_arlib}
install -d -m 0755 %{buildroot}%{_arlog}
install -d -m 0755 %{buildroot}%{_arsysconfdir}
install -d -m 0755 %{buildroot}%{_artmpdir}/pids
install -d -m 0755 %{buildroot}%{_artmpdir}/sockets
install -d -m 0755 %{buildroot}%{_arapps}
install -d -m 0755 %{buildroot}%{_arlog}
install -d -m 0755 %{buildroot}%{_arsystemd}
install -d -m 0755 %{buildroot}%{_arsysconfig}
install -d -m 0755 %{buildroot}%{_arlogrotated}
install -d -m 0755 %{buildroot}%{_arrsyslogd}

cp bin/alert-router %{buildroot}%{_arbin}
cp etc/alert-router.yml %{buildroot}%{_arsysconfdir}
cp etc/alert-router.service %{buildroot}%{_arsystemd}
cp etc/alert-router.sysconfig %{buildroot}%{_arsysconfig}/alert-router
cp etc/logrotate.conf %{buildroot}%{_arlogrotated}/alert-router
cp etc/rsyslog.conf %{buildroot}%{_arrsyslogd}/alert-router.conf

%pre

getent passwd conops &> /dev/null
if [ $? -ne 0 ];then
  echo "The conops user must be created first"
  exit 1
fi 

getent group coed &> /dev/null
if [ $? -ne 0 ];then
  echo "The coed group must be created first"
  exit 1
fi 

%post

systemctl restart rsyslog &> /dev/null || :

if [ $1 -eq 1 ];then
  systemctl enable alert-router &> /dev/null || :
fi

if [ $1 -eq 2 ];then
  # daemon-reload accounts for changes to the service script
  systemctl daemon-reload &> /dev/null || :
  systemctl try-restart &> /dev/null || :
fi

%preun

if [ $1 -eq 0 ];then
  systemctl stop alert-router &> /dev/null || "
  systemctl disable alert-router &> /dev/null || "
fi

%postun

if [ $1 -eq 0 ];then
  systemctl reset-failed &> /dev/null || :
fi

%files
%defattr(-,conops,coed)
%dir %{_arprefix}
%dir %{_arusr}
%dir %{_arvar}
%dir %{_arbin}
%dir %{_arlog}
%dir %{_arsysconfdir}
%dir %{_artmpdir}
%dir %{_artmpdir}/pids
%dir %{_artmpdir}/sockets
%dir %{_arlib}
%dir %{_arapps}
%dir %{_arlog}
%{_arbin}/alert-router
%config %{_arsysconfdir}/alert-router.yml
%{_arsystemd}/alert-router.service
%{_arsysconfig}/alert-router
%{_arlogrotated}/alert-router
%{_arrsyslogd}/alert-router.conf


%changelog
* Mon Dec 10 2018 Greg Land <me@gregland.dev> 0.0.5
initial
