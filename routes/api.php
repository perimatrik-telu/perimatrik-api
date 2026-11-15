<?php

use App\Http\Controllers\ArticleController;
use App\Http\Controllers\AuthController;
use App\Http\Controllers\GalleryController;
use App\Http\Controllers\MemberController;
use App\Http\Controllers\AngkatanController;
use App\Http\Controllers\NewsController;
use App\Http\Controllers\UserController;
use App\Http\Controllers\SettingController;
use App\Http\Controllers\CommentController;
use App\Enums\StatusCode;
use App\Helpers\ResponseHelper;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Route;

Route::prefix('v1')->group(function () {

    // --- Authentication Routes ---
    Route::post('/auth/register', [AuthController::class, 'register']);
    Route::post('/auth/login', [AuthController::class, 'login']);
    Route::middleware('auth:sanctum')->group(function () {
        Route::post('/auth/logout', [AuthController::class, 'logout']);
        Route::post('/auth/change-password', [AuthController::class, 'changePassword']);
        Route::post('/auth/verify-password', [AuthController::class, 'verifyPassword']);
        Route::get('/auth/current-user', [AuthController::class, 'currentUser']);
        Route::get('/user', function (Request $request) {
            return ResponseHelper::success(StatusCode::OK, null, $request->user());
        });
    });

    // --- User Management (Admin Only) ---
    Route::middleware(['auth:sanctum', 'role:admin'])->group(function () {
        Route::get('/system/users', [UserController::class, 'index']);
        Route::post('/system/users', [UserController::class, 'store']);
        Route::delete('/system/users/{id}', [UserController::class, 'destroy']);
    });

    // --- Profile Management (Authenticated) ---
    Route::middleware(['auth:sanctum'])->group(function () {
        // Shared routes for system/users (Logic in controller handles owner vs admin access)
        Route::get('/system/users/{id}', [UserController::class, 'show']);
        Route::put('/system/users/{id}', [UserController::class, 'update']);
        Route::post('/system/users/{id}', [UserController::class, 'update']); // For FormData

        // Legacy / Alternative profile routes
        Route::get('/profile/{id}', [UserController::class, 'show']);
        Route::put('/profile/{id}', [UserController::class, 'update']);
        Route::post('/profile/{id}', [UserController::class, 'update']); 
    });

    // --- Settings ---
    Route::get('/system/settings/contact', [SettingController::class, 'getContactInfo']);
    Route::middleware(['auth:sanctum', 'role:admin'])->group(function () {
        Route::post('/system/settings/contact', [SettingController::class, 'updateContactInfo']);
    });

    // --- Content: Articles ---
    Route::get('/content/article', [ArticleController::class, 'index']);
    Route::get('/content/article/{slug}', [ArticleController::class, 'show']);
    Route::get('/content/article/id/{id}', [ArticleController::class, 'showById']);
    Route::get('/content/article/{slug}/related', [ArticleController::class, 'getRelatedArticles']);

    Route::middleware(['auth:sanctum', 'role:admin,member'])->group(function () {
        Route::post('/content/article', [ArticleController::class, 'store']);
        Route::post('/content/article/{id}', [ArticleController::class, 'update']); // Supports FormData
        Route::put('/content/article/{id}', [ArticleController::class, 'update']);
        Route::delete('/content/article/{id}', [ArticleController::class, 'destroy']);
    });

    // --- Content: News ---
    Route::get('/content/news', [NewsController::class, 'index']);
    Route::get('/content/news/{slug}', [NewsController::class, 'show']);
    Route::get('/content/news/id/{id}', [NewsController::class, 'showById']);
    Route::get('/content/news/{slug}/related', [NewsController::class, 'getRelatedNews']);

    Route::middleware(['auth:sanctum', 'role:admin,officer'])->group(function () {
        Route::post('/content/news', [NewsController::class, 'store']);
        Route::post('/content/news/{id}', [NewsController::class, 'update']);
        Route::put('/content/news/{id}', [NewsController::class, 'update']);
        Route::delete('/content/news/{id}', [NewsController::class, 'destroy']);
    });

    // --- Content: Gallery ---
    Route::get('/content/gallery', [GalleryController::class, 'index']);
    Route::get('/content/gallery/{id}', [GalleryController::class, 'show']);
    Route::get('/images/{path}', [GalleryController::class, 'showImage'])->where('path', '.*')->name('image.show');

    Route::middleware(['auth:sanctum', 'role:admin,officer,member'])->group(function () {
        Route::post('/content/gallery', [GalleryController::class, 'store']);
        Route::post('/content/gallery/{id}', [GalleryController::class, 'update']);
        Route::put('/content/gallery/{id}', [GalleryController::class, 'update']);
        Route::delete('/content/gallery/{id}', [GalleryController::class, 'destroy']);
    });

    // --- Content: Members ---
    Route::get('/content/member', [MemberController::class, 'index']);
    Route::get('/content/export-members', [MemberController::class, 'export']);
    Route::get('/content/member/{id}', [MemberController::class, 'show']);

    Route::middleware(['auth:sanctum', 'role:admin'])->group(function () {
        Route::post('/content/member', [MemberController::class, 'store']);
        Route::post('/content/member/{id}', [MemberController::class, 'update']);
        Route::put('/content/member/{id}', [MemberController::class, 'update']);
        Route::delete('/content/member/{id}', [MemberController::class, 'destroy']);
    });

    // --- Content: Angkatan ---
    Route::get('/content/angkatan', [AngkatanController::class, 'index']);
    Route::get('/content/angkatan/{id}', [AngkatanController::class, 'show']);

    Route::middleware(['auth:sanctum', 'role:admin'])->group(function () {
        Route::post('/content/angkatan', [AngkatanController::class, 'store']);
        Route::put('/content/angkatan/{id}', [AngkatanController::class, 'update']);
        Route::delete('/content/angkatan/{id}', [AngkatanController::class, 'destroy']);
    });

    // --- Comments ---
    Route::get('/content/comments', [CommentController::class, 'index']);
    Route::middleware(['auth:sanctum'])->group(function () {
        Route::post('/content/comments', [CommentController::class, 'store']);
        Route::delete('/content/comments/{id}', [CommentController::class, 'destroy']);
    });

    // --- System Status / Test Routes ---
    Route::middleware(['auth:sanctum', 'role:admin'])->group(function () {
        Route::get('/admin-only', function () {
            return ResponseHelper::success(StatusCode::OK, 'Welcome, Admin!');
        });
    });
});
