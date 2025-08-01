## 为何不使用 Go-Vips
GO VIPS 在底层调用依然需要 C 工具的参与，而这些工具在 Windows 中的编译和使用都存在问题。尤其是对于 Windows 用户来说，安装和配置这些工具可能会非常麻烦。
只有在 Mac 或者 Linux 系统中，使用 GO VIPS 才是一个不错的选择。
故发行版中不提供 GO VIPS 的支持。未来可能会考虑提供一个基于 Docker 的解决方案，使用 GOVips 来进行性能提升.