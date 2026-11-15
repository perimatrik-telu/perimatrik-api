<?php

namespace Database\Seeders;

use App\Models\News;
use App\Models\User;
use Illuminate\Database\Seeder;
use Illuminate\Support\Str;

class NewsSeeder extends Seeder
{
    /**
     * Run the database seeds.
     */
    public function up(): void
    {
        $user = User::first();
        if (!$user) return;

        for ($i = 1; $i <= 5; $i++) {
            $title = "Berita PERIMATRIK Ke-" . $i;
            News::create([
                'title' => $title,
                'slug' => Str::slug($title) . '-' . time() . '-' . $i,
                'body' => '<p>Ini adalah konten berita contoh untuk berita ke-' . $i . '. PERIMATRIK terus bergerak maju!</p>',
                'image' => 'https://picsum.photos/seed/news' . $i . '/800/600',
                'status' => 'active',
                'user_id' => $user->id,
                'published_at' => now(),
            ]);
        }
    }
    
    /**
     * Run the database seeds.
     */
    public function run(): void
    {
        $this->up();
    }
}
