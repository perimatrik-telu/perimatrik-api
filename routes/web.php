<?php

use Illuminate\Support\Facades\Route;
use Illuminate\Support\Str;
use App\Models\Article;
use App\Models\News;

Route::get('/', function () {
    return response()->json(['message' => 'Hello World']);
});

Route::get('/share/article/{slug}', function ($slug) {
    $article = Article::where('slug', $slug)->firstOrFail();
    $frontendUrl = env('FRONTEND_URL', 'http://localhost:5173');
    
    return view('share', [
        'title' => $article->title,
        'description' => Str::limit(strip_tags($article->body ?? ''), 150),
        'image' => $article->image,
        'url' => rtrim($frontendUrl, '/') . '/article/detail/' . $article->slug
    ]);
});

Route::get('/share/news/{slug}', function ($slug) {
    $news = News::where('slug', $slug)->firstOrFail();
    $frontendUrl = env('FRONTEND_URL', 'http://localhost:5173');
    
    return view('share', [
        'title' => $news->title,
        'description' => Str::limit(strip_tags($news->body ?? ''), 150),
        'image' => $news->image,
        'url' => rtrim($frontendUrl, '/') . '/news/detail/' . $news->slug
    ]);
});
