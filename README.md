# virt-tui
virt-tuiはlibvirtや操作VMの状態確認を行えるTUIアプリケーションです。  
virt-tuiは[virt-manager](https://github.com/virt-manager/virt-manager)のように簡単にlibvirtを操作できることを目指して作成されています。
(現時点ではほとんど出来ていません。)

virt-tui is a TUI application that allows you to check the status of libvirt and operating VMs.  
virt-tui was created with the goal of making it as easy to manipulate libvirt as [virt-manager](https://github.com/virt-manager/virt-manager). (At the moment, it is almost not done.)

## 前提導入パッケージ (Prerequisite Installation Package)
#### Ubuntu 20.04
``` bash
sudo apt install qemu-kvm qemu-system libvirt-daemon-system libvirt-daemon libvirt-dev libvirt-clients bridge-utils libosinfo-bin libguestfs-tools virt-top cloud-image-utils virtinst
```  
go version 1.18

## Install virt-tui
``` bash
piyo@ubuntu:~$ id
uid=1000(piyo) gid=1000(piyo) groups=1000(piyo),4(adm),24(cdrom),27(sudo),30(dip),46(plugdev),120(lpadmin),131(sambashare)
```
ユーザーにlibvirtの権限を付与します。  
Grant libvirt privileges to users.
``` bash
piyo@ubuntu-hoge:~$ sudo gpasswd -a piyo libvirt
sudo gpasswd -a piyo libvirt
Adding user piyo to group libvirt
piyo@ubuntu-hoge:~$ id
uid=1000(piyo) gid=1000(piyo) groups=1000(piyo),4(adm),24(cdrom),27(sudo),30(dip),46(plugdev),120(lpadmin),131(sambashare),134(libvirt)
```
Install  
``` bash
git clone https://github.com/nyanco01/virt-tui
cd virt-tui  
go build
./virt-tui
```

## 備考 (note)
メモリ使用率はVMのxmlファイルのmemballonを以下のように`<stats period='1'/>`を追加しないと正しく機能しません。
```xml
    <memballoon model='virtio'>
      <stats period='1'/>
      <alias name='balloon0'/>
    </memballoon>
```
