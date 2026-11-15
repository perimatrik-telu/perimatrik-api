<?php
namespace App\Http\Controllers;

use App\Enums\ResponseMessage;
use App\Enums\StatusCode;
use App\Helpers\AttachmentHelper;
use App\Helpers\ResponseHelper;
use App\Models\Article;
use App\Models\User;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\Log;       // Import Log facade
use Illuminate\Support\Facades\Validator; // Import User model
use Illuminate\Support\Str;

// Import UploadedFile class

class ArticleController extends Controller
{
    /**
     * Display a listing of the resource.
     */
    public function index(Request $request)
    {
        $page   = $request->input('page', 1);
        $limit  = $request->input('limit', 10);
        $filter = $request->input('filter', "DESC");
        $userId = $request->input('user_id'); 
        $skip   = ($page - 1) * $limit;

        Log::info("Article index called with user_id: " . ($userId ?? 'null'));

        $authUser = Auth::guard('sanctum')->user();
        $query    = Article::with('user');

        if ($userId && $userId !== 'undefined') {
            if (! $authUser || ($authUser->role !== 'admin' && $authUser->id !== $userId)) {
                return ResponseHelper::responseMessageRequest(StatusCode::FORBIDDEN, false, ResponseMessage::UNAUTHORIZED);
            }
            $query->where('user_id', $userId);
        } else {
            if (! $authUser || $authUser->role !== 'admin') {
                $query->where('status', 'active')
                    ->whereNotNull('published_at')
                    ->where('published_at', '<=', now());
            }
        }

        $total = (clone $query)->count();
        $data  = $query->skip($skip)
            ->take($limit)
            ->orderBy('published_at', $filter)
            ->get();

        return ResponseHelper::responseSuccessDataWithPaginate($data, $page, $limit, $total, StatusCode::OK);
    }

    /**
     * Show the form for creating a new resource.
     */
    public function create()
    {
        //
    }

    /**
     * Store a newly created resource in storage.
     */
    public function store(Request $request)
    {
        $user = Auth::user();
        $isDraft = $request->input('status') === 'draft';

        $rules = [
            'title'        => $isDraft ? 'nullable|string' : 'required|string',
            'body'         => $isDraft ? 'nullable|string' : 'required|string',
            'published_at' => 'nullable|date', // Add validation rule
            'status'       => 'nullable|in:draft,active,inactive',
        ];

        if ($request->hasFile('image')) {
            $rules['image'] = 'nullable|file|mimes:jpeg,png,jpg,gif,svg|max:2048';
        } else {
            $rules['image'] = $isDraft ? 'nullable|string' : 'required|string|url';
        }

        $validator = Validator::make($request->all(), $rules);

        if ($validator->fails()) {
            return ResponseHelper::responseMessageRequestWithErrors(StatusCode::BAD_REQUEST, false, ResponseMessage::FAILED_CREATE, $validator->errors());
        }

        $fullImageUrl = null;
        if ($request->hasFile('image')) {
            $filename     = AttachmentHelper::saveImageMedia($request, 'image', config('path.ARTICLE_PATH'));
            $fullImageUrl = config('app.url') . '/storage/' . ltrim(config('path.ARTICLE_PATH'), '/') . '/' . $filename;
        } else {
            $fullImageUrl = $request->input('image');
        }

        $title = $request->title ?? 'Untitled Draft';
        $slug = Str::slug($title, '-') . '-' . time();

        $article               = new Article();
        $article->slug         = $slug;
        $article->image        = $fullImageUrl ?? '';
        $article->title        = $title;
        $article->body         = $request->body ?? '';
        $article->user_id      = $user->id;
        $article->status       = $request->input('status', 'active');
        $article->published_at = $article->status === 'active'
            ? ($request->published_at ?? now())
            : null;
        $article->save();

        return ResponseHelper::responseMessageRequest(StatusCode::CREATED, true, ResponseMessage::SUCCESS_CREATE);
    }

    /**
     * Display the specified resource.
     */
    public function show($slug)
    {
        $article = Article::where('slug', $slug)->with('user')->first();
        if (! $article) {
            return ResponseHelper::responseMessageRequest(StatusCode::NOT_FOUND, false, ResponseMessage::NO_ITEM);
        }

        if (! $this->canViewArticle($article)) {
            return ResponseHelper::responseMessageRequest(StatusCode::NOT_FOUND, false, ResponseMessage::NO_ITEM);
        }

        $article->increment('views');

        $article->load('user'); // Eager load the user relationship

        // Append author and postingAt attributes
        $article->append([
            'author',
            'postingAt',
        ]);

        return ResponseHelper::responseSuccessSingleData($article, StatusCode::OK);
    }

    /**
     * Display the specified resource by ID.
     */
    public function showById($id)
    {
        $article = Article::find($id); // Find by ID
        if (! $article) {
            return ResponseHelper::responseMessageRequest(StatusCode::NOT_FOUND, false, ResponseMessage::NO_ITEM);
        }

        if (! $this->canViewArticle($article)) {
            return ResponseHelper::responseMessageRequest(StatusCode::NOT_FOUND, false, ResponseMessage::NO_ITEM);
        }

        $article->increment('views'); // Still increment views for a view by ID

        $article->load('user'); // Eager load the user relationship

        // Append author and postingAt attributes
        $article->append([
            'author',
            'postingAt',
        ]);

        return ResponseHelper::responseSuccessSingleData($article, StatusCode::OK);
    }

    /**
     * Get related articles, excluding the current one.
     */
    public function getRelatedArticles($slug)
    {
        $limit = 3; // You can adjust the limit as needed

        $data = Article::where('slug', '!=', $slug)
            ->where('status', 'active')
            ->whereNotNull('published_at')
            ->where('published_at', '<=', now())
            ->with('user')
            ->orderBy('created_at', 'desc')
            ->take($limit)
            ->get();

        return ResponseHelper::responseSuccessSingleData($data, StatusCode::OK);
    }

    /**
     * Show the form for editing the specified resource.
     */
    public function edit(Article $article)
    {
        //
    }

    /**
     * Update the specified resource in storage.
     */
    public function update(Request $request, $id)
    {
        $article = Article::find($id);
        if (! $article) {
            return ResponseHelper::responseMessageRequest(StatusCode::NOT_FOUND, false, ResponseMessage::NO_ITEM);
        }

        $isDraft = $request->input('status') === 'draft' || ($article->status === 'draft' && !$request->has('status'));

        $rules = [
            'title'        => $isDraft ? 'nullable|string' : 'required|string',
            'body'         => $isDraft ? 'nullable|string' : 'required|string',
            'image'        => 'nullable|file|mimes:jpeg,png,jpg,gif,svg|max:2048',
            'published_at' => 'nullable|date',
            'status'       => 'nullable|in:draft,active,inactive',
        ];

        $dataToValidate = $request->only(['title', 'body', 'published_at', 'status']);
        if ($request->hasFile('image')) {
            $dataToValidate['image'] = $request->file('image');
        }

        $validator = Validator::make($dataToValidate, $rules);

        if ($validator->fails()) {
            return ResponseHelper::responseMessageRequestWithErrors(
                StatusCode::BAD_REQUEST,
                false,
                ResponseMessage::FAILED_UPDATE,
                $validator->errors()
            );
        }

        if ($request->hasFile('image')) {
            // Delete old image if it exists
            if ($article->image) {
                AttachmentHelper::deleteMedia(config('path.ARTICLE_PATH'), basename(parse_url($article->image, PHP_URL_PATH)));
            }
            // Save new image
            $filename       = AttachmentHelper::saveImageMedia($request, 'image', config('path.ARTICLE_PATH'));
            $article->image = config('app.url') . '/storage/' . ltrim(config('path.ARTICLE_PATH'), '/') . '/' . $filename;
        }

        // Generate and update slug if title has changed
        $newTitle = $request->title ?? 'Untitled Draft';
        if ($request->has('title') && $newTitle !== $article->title) {
            $article->slug = Str::slug($newTitle, '-') . '-' . time();
        }

        if ($request->has('title')) {
            $article->title = $newTitle;
        }
        
        if ($request->has('body') || $isDraft) {
            $article->body  = $request->body ?? '';
        }

        if ($request->has('status')) {
            $article->status = $request->input('status');
        }

        if ($article->status === 'draft') {
            $article->published_at = null;
        } elseif ($request->has('published_at')) {
            $article->published_at = $request->filled('published_at') ? $request->published_at : now();
        } elseif (! $article->published_at) {
            $article->published_at = now();
        }

        $article->save();

        return ResponseHelper::responseMessageRequest(
            StatusCode::OK,
            true,
            ResponseMessage::SUCCESS_UPDATE
        );
    }

    /**
     * Remove the specified resource from storage.
     */
    public function destroy($id)
    {
        $article = Article::find($id);
        if (! $article) {
            Log::warning("Attempted to delete a non-existent article with ID: {$id}.");
            return ResponseHelper::responseMessageRequest(StatusCode::NOT_FOUND, false, ResponseMessage::NO_ITEM);
        }

        Log::info("Attempting to delete article with ID: {$article->id} and slug: {$article->slug}");

        // Extract filename from the full image URL before deleting
        if ($article->image && str_contains($article->image, '/storage/')) {
            AttachmentHelper::deleteMedia(config('path.ARTICLE_PATH'), basename(parse_url($article->image, PHP_URL_PATH)));
        }

        $deleted = $article->delete();

        if ($deleted) {
            Log::info("Article with ID: {$article->id} and slug: {$article->slug} successfully deleted.");
            return ResponseHelper::responseMessageRequest(StatusCode::OK, true, ResponseMessage::SUCCESS_DELETE);
        } else {
            Log::error("Failed to delete article with ID: {$article->id} and slug: {$article->slug}.");
            return ResponseHelper::responseMessageRequest(StatusCode::INTERNAL_SERVER_ERROR, false, "Failed to delete article.");
        }
    }

    private function canViewArticle(Article $article): bool
    {
        if ($article->status === 'active' && $article->published_at && $article->published_at <= now()) {
            return true;
        }

        $user = Auth::guard('sanctum')->user();

        return $user && ($user->role === 'admin' || $user->id === $article->user_id);
    }
}
