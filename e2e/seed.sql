SET @pw = '$2y$12$n/.LKWeKfX/eQfFgE02.IuXGo8oUTUiwab6o1668ax/rAZ7IQtQlK';
INSERT INTO users (id,name,email,password,role,avatar,created_at,updated_at) VALUES
('a0000000-0000-4000-8000-000000000001','Admin E2E','admin@e2e.test',@pw,'admin',NULL,'2025-01-01 00:00:00','2025-01-01 00:00:00'),
('a0000000-0000-4000-8000-000000000002','Officer E2E','officer@e2e.test',@pw,'officer',NULL,'2025-01-02 00:00:00','2025-01-02 00:00:00'),
('a0000000-0000-4000-8000-000000000003','Budi Santoso','budi@e2e.test',@pw,'member','http://api.test/storage/avatars/budi.png','2025-01-03 00:00:00','2025-01-03 00:00:00'),
('a0000000-0000-4000-8000-000000000004','Lain E2E','lain@e2e.test',@pw,'member',NULL,'2025-01-04 00:00:00','2025-01-04 00:00:00');
INSERT INTO angkatans VALUES
('b0000000-0000-4000-8000-000000000001','E2E-ANGKATAN XX','2020','2025-01-01 00:00:00','2025-01-01 00:00:00'),
('b0000000-0000-4000-8000-000000000002','Angkatan XXI','2021','2025-01-01 00:00:00','2025-01-01 00:00:00');
INSERT INTO members (id,user_id,no_anggota,nama_lengkap,nama_lapangan,domisili,kontak,avatar,angkatan_id,created_at,updated_at,instagram) VALUES
('c0000000-0000-4000-8000-000000000001','a0000000-0000-4000-8000-000000000003','001','Budi Santoso','Badak','Bandung','0811','http://api.test/storage/avatars/budi.png','b0000000-0000-4000-8000-000000000001','2025-01-01 00:00:00','2025-01-01 00:00:00','@budi'),
('c0000000-0000-4000-8000-000000000002',NULL,'002','Siti Aminah',NULL,'Jakarta',NULL,NULL,'b0000000-0000-4000-8000-000000000002','2025-01-02 00:00:00','2025-01-02 00:00:00',NULL),
('c0000000-0000-4000-8000-000000000003',NULL,'0123','Andi "Si Cepat"','Elang','Bogor','08123',NULL,'b0000000-0000-4000-8000-000000000001','2025-01-03 00:00:00','2025-01-03 00:00:00',NULL);
INSERT INTO articles (id,slug,image,title,body,views,status,published_at,user_id,created_at,updated_at) VALUES
('d0000000-0000-4000-8000-000000000001','artikel-pertama-1700000000','http://api.test/storage/articles/seed.jpg','Artikel Pertama','<p>Isi <b>artikel</b> pertama yang cukup panjang untuk deskripsi share. Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.</p>',5,'active','2025-01-10 08:00:00','a0000000-0000-4000-8000-000000000001','2025-01-10 07:00:00','2025-01-10 07:00:00'),
('d0000000-0000-4000-8000-000000000002','draft-budi-1700000001','','Draft Budi','draft',0,'draft',NULL,'a0000000-0000-4000-8000-000000000003','2025-01-11 07:00:00','2025-01-11 07:00:00'),
('d0000000-0000-4000-8000-000000000003','masa-depan-1700000002','https://example.com/x.jpg','Masa Depan','future',0,'active','2030-01-01 00:00:00','a0000000-0000-4000-8000-000000000003','2025-01-12 07:00:00','2025-01-12 07:00:00'),
('d0000000-0000-4000-8000-000000000004','artikel-budi-1700000003','https://example.com/y.jpg','Artikel Budi','<h1>Halo</h1>',2,'active','2025-02-01 09:30:00','a0000000-0000-4000-8000-000000000003','2025-01-13 07:00:00','2025-01-13 07:00:00');
INSERT INTO news (id,slug,image,title,body,views,status,published_at,user_id,created_at,updated_at) VALUES
('e0000000-0000-4000-8000-000000000001','berita-satu-1700000000','http://api.test/storage/news/seed.jpg','Berita Satu','<p>Berita</p>',1,'draft','2025-01-05 10:00:00','a0000000-0000-4000-8000-000000000001','2025-01-05 10:00:00','2025-01-05 10:00:00'),
('e0000000-0000-4000-8000-000000000002','berita-dua-1700000001',NULL,'Berita Dua','dua',0,'active','2025-03-01 10:00:00','a0000000-0000-4000-8000-000000000002','2025-03-01 10:00:00','2025-03-01 10:00:00');
INSERT INTO galleries (id,image,caption,slug,user_id,created_at,updated_at,latitude,longitude) VALUES
('f0000000-0000-4000-8000-000000000001','http://api.test/storage/galleries/e2e-seed.png','Foto Seed','foto-seed','a0000000-0000-4000-8000-000000000003','2025-01-01 00:00:00','2025-01-01 00:00:00',-6.2,106.8),
('f0000000-0000-4000-8000-000000000002','https://example.com/g.jpg',NULL,'abc','a0000000-0000-4000-8000-000000000001','2025-01-02 00:00:00','2025-01-02 00:00:00',NULL,NULL);
INSERT INTO settings VALUES
('99000000-0000-4000-8000-000000000001','contact_email','info@e2e.test','2025-01-01 00:00:00','2025-01-01 00:00:00'),
('99000000-0000-4000-8000-000000000002','contact_maps',NULL,'2025-01-01 00:00:00','2025-01-01 00:00:00');
INSERT INTO comments VALUES
('70000000-0000-4000-8000-000000000001','Komentar pertama','a0000000-0000-4000-8000-000000000003','d0000000-0000-4000-8000-000000000001',NULL,NULL,'2025-02-01 00:00:00','2025-02-01 00:00:00'),
('70000000-0000-4000-8000-000000000002','Balasan admin','a0000000-0000-4000-8000-000000000001','d0000000-0000-4000-8000-000000000001',NULL,'70000000-0000-4000-8000-000000000001','2025-02-02 00:00:00','2025-02-02 00:00:00'),
('70000000-0000-4000-8000-000000000003','Balasan dari balasan','a0000000-0000-4000-8000-000000000002',NULL,NULL,'70000000-0000-4000-8000-000000000002','2025-02-03 00:00:00','2025-02-03 00:00:00'),
('70000000-0000-4000-8000-000000000004','Komentar officer','a0000000-0000-4000-8000-000000000002','d0000000-0000-4000-8000-000000000001',NULL,NULL,'2025-02-04 00:00:00','2025-02-04 00:00:00');
