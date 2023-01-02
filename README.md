# virt-tui
<img src="https://user-images.githubusercontent.com/112300116/189452942-6abe3c3d-87e6-4c88-ada1-3e9fa54c3900.png" width="600">
virt-tuiはlibvirtや操作VMの状態確認を行えるTUIアプリケーションです。  
virt-tuiはvirt-managerのように簡単にlibvirtを操作できることを目指して作成されています。(現時点ではほとんど出来ていません。)

virt-tui is a TUI application that allows you to check the status of libvirt and operating VMs.  
virt-tui was created with the goal of making it as easy to manipulate libvirt as [virt-manager](https://github.com/virt-manager/virt-manager). (At the moment, it is almost not done.)

## 前提導入パッケージ (Prerequisite Installation Package)
#### Ubuntu 20.04 || Ubuntu 22.04
``` bash
sudo apt install qemu-kvm qemu-system libvirt-daemon-system libvirt-daemon libvirt-dev libvirt-clients bridge-utils libosinfo-bin libguestfs-tools virt-top cloud-image-utils virtinst
```  
go version 1.18
[How to Install](https://go.dev/doc/install)

## Install virt-tui
Ubuntu22.04のみ`/etc/libvirt/qemu.conf`を以下のように設定してください。
```

Install  
``` bash
git clone https://github.com/nyanco01/virt-tui
cd virt-tui  
go build
sudo ./virt-tui
```
※`sudo`なしではVMの作成ができません。  
VM cannot be created without sudo.

## 備考 (note)
メモリの使用率を正しく表示させる為には、VMを定義しているxmlファイル内のmemballonの項目に`<stats period='1'/>`を追加する必要があります。  
In order to display the memory usage correctly, you need to add an `<stats period='1'/>` to the memballon entry in the xml file that defines the VM. 
```xml
    <memballoon model='virtio'>
      <stats period='1'/>
      <alias name='balloon0'/>
    </memballoon>
```
