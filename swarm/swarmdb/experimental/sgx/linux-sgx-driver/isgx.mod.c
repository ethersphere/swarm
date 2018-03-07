#include <linux/module.h>
#include <linux/vermagic.h>
#include <linux/compiler.h>

MODULE_INFO(vermagic, VERMAGIC_STRING);

struct module __this_module
__attribute__((section(".gnu.linkonce.this_module"))) = {
	.name = KBUILD_MODNAME,
	.init = init_module,
#ifdef CONFIG_MODULE_UNLOAD
	.exit = cleanup_module,
#endif
	.arch = MODULE_ARCH_INIT,
};

static const struct modversion_info ____versions[]
__used
__attribute__((section("__versions"))) = {
	{ 0x28950ef1, __VMLINUX_SYMBOL_STR(module_layout) },
	{ 0x8487a2b6, __VMLINUX_SYMBOL_STR(flush_work) },
	{ 0x434693dd, __VMLINUX_SYMBOL_STR(alloc_pages_current) },
	{ 0x98ab5c8d, __VMLINUX_SYMBOL_STR(kmalloc_caches) },
	{ 0x65e75cb6, __VMLINUX_SYMBOL_STR(__list_del_entry) },
	{ 0x36fd4a13, __VMLINUX_SYMBOL_STR(put_pid) },
	{ 0xf5893abf, __VMLINUX_SYMBOL_STR(up_read) },
	{ 0x4c4fef19, __VMLINUX_SYMBOL_STR(kernel_stack) },
	{ 0x784213a6, __VMLINUX_SYMBOL_STR(pv_lock_ops) },
	{ 0x4b7dcf38, __VMLINUX_SYMBOL_STR(_raw_qspin_lock) },
	{ 0x43a53735, __VMLINUX_SYMBOL_STR(__alloc_workqueue_key) },
	{ 0xc8b57c27, __VMLINUX_SYMBOL_STR(autoremove_wake_function) },
	{ 0x45449b56, __VMLINUX_SYMBOL_STR(boot_cpu_data) },
	{ 0xf3ba11d9, __VMLINUX_SYMBOL_STR(mmu_notifier_register) },
	{ 0xf087137d, __VMLINUX_SYMBOL_STR(__dynamic_pr_debug) },
	{ 0x4ed12f73, __VMLINUX_SYMBOL_STR(mutex_unlock) },
	{ 0xb862e2ca, __VMLINUX_SYMBOL_STR(platform_device_register_full) },
	{ 0x97651e6c, __VMLINUX_SYMBOL_STR(vmemmap_base) },
	{ 0x120eee66, __VMLINUX_SYMBOL_STR(set_page_dirty) },
	{ 0xc35e4b4e, __VMLINUX_SYMBOL_STR(kthread_create_on_node) },
	{ 0x5b0240c6, __VMLINUX_SYMBOL_STR(__platform_driver_register) },
	{ 0x343a1a8, __VMLINUX_SYMBOL_STR(__list_add) },
	{ 0x57a6ccd0, __VMLINUX_SYMBOL_STR(down_read) },
	{ 0x71de9b3f, __VMLINUX_SYMBOL_STR(_copy_to_user) },
	{ 0xf23b2e74, __VMLINUX_SYMBOL_STR(misc_register) },
	{ 0xb8c7ff88, __VMLINUX_SYMBOL_STR(current_task) },
	{ 0x9a025cd5, __VMLINUX_SYMBOL_STR(__mutex_init) },
	{ 0x27e1a049, __VMLINUX_SYMBOL_STR(printk) },
	{ 0x42f90a31, __VMLINUX_SYMBOL_STR(kthread_stop) },
	{ 0x6283e9a7, __VMLINUX_SYMBOL_STR(apply_to_page_range) },
	{ 0x94cbf6d4, __VMLINUX_SYMBOL_STR(shmem_read_mapping_page_gfp) },
	{ 0xa1c76e0a, __VMLINUX_SYMBOL_STR(_cond_resched) },
	{ 0x7d5a2c73, __VMLINUX_SYMBOL_STR(on_each_cpu_mask) },
	{ 0x16305289, __VMLINUX_SYMBOL_STR(warn_slowpath_null) },
	{ 0x9abdea30, __VMLINUX_SYMBOL_STR(mutex_lock) },
	{ 0x8c03d20c, __VMLINUX_SYMBOL_STR(destroy_workqueue) },
	{ 0x521445b, __VMLINUX_SYMBOL_STR(list_del) },
	{ 0x9b36baa6, __VMLINUX_SYMBOL_STR(platform_device_unregister) },
	{ 0x7fbd10d2, __VMLINUX_SYMBOL_STR(radix_tree_next_chunk) },
	{ 0xcd1850f2, __VMLINUX_SYMBOL_STR(fput) },
	{ 0x200df995, __VMLINUX_SYMBOL_STR(shmem_file_setup) },
	{ 0x3c483012, __VMLINUX_SYMBOL_STR(radix_tree_delete) },
	{ 0xf11543ff, __VMLINUX_SYMBOL_STR(find_first_zero_bit) },
	{ 0x7cd8d75e, __VMLINUX_SYMBOL_STR(page_offset_base) },
	{ 0xfeb44a7, __VMLINUX_SYMBOL_STR(find_vma) },
	{ 0xa281370e, __VMLINUX_SYMBOL_STR(__free_pages) },
	{ 0x62fdffd1, __VMLINUX_SYMBOL_STR(mmu_notifier_unregister_no_release) },
	{ 0xf0fdf6cb, __VMLINUX_SYMBOL_STR(__stack_chk_fail) },
	{ 0x1000e51, __VMLINUX_SYMBOL_STR(schedule) },
	{ 0x79a38e61, __VMLINUX_SYMBOL_STR(___ratelimit) },
	{ 0x910538ff, __VMLINUX_SYMBOL_STR(pv_cpu_ops) },
	{ 0xe65cdceb, __VMLINUX_SYMBOL_STR(wake_up_process) },
	{ 0xbdfb6dbb, __VMLINUX_SYMBOL_STR(__fentry__) },
	{ 0xcc5005fe, __VMLINUX_SYMBOL_STR(msleep_interruptible) },
	{ 0x41ec4c1a, __VMLINUX_SYMBOL_STR(kmem_cache_alloc_trace) },
	{ 0xcf21d241, __VMLINUX_SYMBOL_STR(__wake_up) },
	{ 0xb3f7646e, __VMLINUX_SYMBOL_STR(kthread_should_stop) },
	{ 0x4c83a22c, __VMLINUX_SYMBOL_STR(zap_vma_ptes) },
	{ 0x37a0cba, __VMLINUX_SYMBOL_STR(kfree) },
	{ 0x69acdf38, __VMLINUX_SYMBOL_STR(memcpy) },
	{ 0x5c8b5ce8, __VMLINUX_SYMBOL_STR(prepare_to_wait) },
	{ 0x72392fe6, __VMLINUX_SYMBOL_STR(pv_mmu_ops) },
	{ 0xedc03953, __VMLINUX_SYMBOL_STR(iounmap) },
	{ 0x68c7263, __VMLINUX_SYMBOL_STR(ioremap_cache) },
	{ 0x334c1f75, __VMLINUX_SYMBOL_STR(put_page) },
	{ 0xfa66f77c, __VMLINUX_SYMBOL_STR(finish_wait) },
	{ 0x844e3767, __VMLINUX_SYMBOL_STR(radix_tree_lookup) },
	{ 0x2e0d2f7f, __VMLINUX_SYMBOL_STR(queue_work_on) },
	{ 0x79a66c5a, __VMLINUX_SYMBOL_STR(platform_driver_unregister) },
	{ 0x77e2f33, __VMLINUX_SYMBOL_STR(_copy_from_user) },
	{ 0x614bb773, __VMLINUX_SYMBOL_STR(radix_tree_insert) },
	{ 0xa1012e43, __VMLINUX_SYMBOL_STR(misc_deregister) },
	{ 0xdf289dfe, __VMLINUX_SYMBOL_STR(vm_insert_pfn) },
};

static const char __module_depends[]
__used
__attribute__((section(".modinfo"))) =
"depends=";

MODULE_ALIAS("acpi*:INT0E0C:*");

MODULE_INFO(srcversion, "695569D1914239D34CC8074");
MODULE_INFO(rhelversion, "7.4");
