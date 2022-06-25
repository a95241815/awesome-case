CREATE TABLE `sail_upload` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `shop_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '商店唯一ID',
  `file_md5` varchar(255) NOT NULL DEFAULT '' COMMENT '文件md5',
  `file_key` varchar(255) NOT NULL DEFAULT '' COMMENT 'Oss FileKey',
  `file_key1` varchar(255) NOT NULL DEFAULT '' COMMENT 'Oss FileKey 750',
  `file_key2` varchar(255) NOT NULL DEFAULT '' COMMENT 'Oss FileKey 900',
  `file_key3` varchar(255) NOT NULL DEFAULT '' COMMENT 'Oss FileKey 1080',
  `image_width` int(25) NOT NULL DEFAULT '0' COMMENT '图片实际宽度',
  `is_del` tinyint(3) NOT NULL DEFAULT '0' COMMENT '是否删除',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `shop_id_index` (`shop_id`),
  KEY `idx_shop_file_md5` (`shop_id`,`file_md5`),
  KEY `idx_file_md5` (`file_md5`(20))
) ENGINE=InnoDB AUTO_INCREMENT=1429906 DEFAULT CHARSET=utf8mb4 COMMENT='图片总表';