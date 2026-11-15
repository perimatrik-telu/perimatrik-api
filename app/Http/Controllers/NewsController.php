<?php

namespace App\Http\Controllers;

use App\Enums\ResponseMessage;
use App\Enums\StatusCode;
use App\Helpers\AttachmentHelper;
use App\Helpers\ResponseHelper;
use App\Models\News;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Validator;
use Illuminate\Support\Str;

class NewsController extends Controller
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

        Log::info("News index called with user_id: " . ($userId ?? 'null'));

        $query = News::with('user');

        if ($userId && $userId !== 'undefined') {
            $query->where('user_id', $userId);
        }

        $total = (clone $query)->count();
        $data  = $query->skip($skip)
            ->take($limit)
            ->orderBy('published_at', $filter)
            ->get();

        return ResponseHelper::responseSuccessDataWithPaginate($data, $page, $limit, $total, StatusCode::OK);
    }

    /**
     * Store a newly created resource in storage.
     */
    public function store(Request $request)
    {
        $user = Auth::user();

        $rules = [
            'title'        => 'required|string',
            'body'         => 'required|string',
            'published_at' => 'nullable|date',
        ];

        if ($request->hasFile('image')) {
            $rules['image'] = 'required|file|mimes:jpeg,png,jpg,gif,svg|max:2048';
        } else {
            $rules['image'] = 'required|string|url';
        }

        $validator = Validator::make($request->all(), $rules);

        if ($validator->fails()) {
            return ResponseHelper::responseMessageRequestWithErrors(StatusCode::BAD_REQUEST, false, ResponseMessage::FAILED_CREATE, $validator->errors());
        }

        $fullImageUrl = null;
        if ($request->hasFile('image')) {
            $filename     = AttachmentHelper::saveImageMedia($request, 'image', config('path.NEWS_PATH'));
            $fullImageUrl = config('app.url') . '/storage/' . ltrim(config('path.NEWS_PATH'), '/') . '/' . $filename;
        } else {
            $fullImageUrl = $request->input('image');
        }

        $slug = Str::slug($request->title, '-') . '-' . time();

        $news               = new News();
        $news->slug         = $slug;
        $news->image        = $fullImageUrl;
        $news->title        = $request->title;
        $news->body         = $request->body;
        $news->user_id      = $user->id;
        $news->published_at = $request->published_at ?? now();
        $news->save();

        return ResponseHelper::responseMessageRequest(StatusCode::CREATED, true, ResponseMessage::SUCCESS_CREATE);
    }

    /**
     * Display the specified resource.
     */
    public function show($slug)
    {
        $news = News::where('slug', $slug)->with('user')->first();
        if (! $news) {
            return ResponseHelper::responseMessageRequest(StatusCode::NOT_FOUND, false, ResponseMessage::NO_ITEM);
        }

        $news->increment('views');
        $news->load('user');

        $news->append([
            'author',
            'postingAt',
        ]);

        return ResponseHelper::responseSuccessSingleData($news, StatusCode::OK);
    }

    /**
     * Display the specified resource by ID.
     */
    public function showById($id)
    {
        $news = News::find($id);
        if (! $news) {
            return ResponseHelper::responseMessageRequest(StatusCode::NOT_FOUND, false, ResponseMessage::NO_ITEM);
        }

        $news->increment('views');
        $news->load('user');

        $news->append([
            'author',
            'postingAt',
        ]);

        return ResponseHelper::responseSuccessSingleData($news, StatusCode::OK);
    }

    /**
     * Get related news, excluding the current one.
     */
    public function getRelatedNews($slug)
    {
        $limit = 3;

        $data = News::where('slug', '!=', $slug)
            ->with('user')
            ->orderBy('created_at', 'desc')
            ->take($limit)
            ->get();

        return ResponseHelper::responseSuccessSingleData($data, StatusCode::OK);
    }

    /**
     * Update the specified resource in storage.
     */
    public function update(Request $request, $id)
    {
        $news = News::find($id);
        if (! $news) {
            return ResponseHelper::responseMessageRequest(StatusCode::NOT_FOUND, false, ResponseMessage::NO_ITEM);
        }

        $rules = [
            'title'        => 'required|string',
            'body'         => 'required|string',
            'image'        => 'nullable|file|mimes:jpeg,png,jpg,gif,svg|max:2048',
            'published_at' => 'nullable|date',
        ];

        $dataToValidate = $request->only(['title', 'body', 'published_at']);
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
            if ($news->image && str_contains($news->image, '/storage/')) {
                AttachmentHelper::deleteMedia(config('path.NEWS_PATH'), basename(parse_url($news->image, PHP_URL_PATH)));
            }
            $filename    = AttachmentHelper::saveImageMedia($request, 'image', config('path.NEWS_PATH'));
            $news->image = config('app.url') . '/storage/' . ltrim(config('path.NEWS_PATH'), '/') . '/' . $filename;
        }

        if ($request->has('title') && $request->title !== $news->title) {
            $news->slug = Str::slug($request->title, '-') . '-' . time();
        }

        $news->title = $request->title;
        $news->body  = $request->body;
        if ($request->has('published_at')) {
            $news->published_at = $request->published_at;
        }
        $news->save();

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
        $news = News::find($id);
        if (! $news) {
            return ResponseHelper::responseMessageRequest(StatusCode::NOT_FOUND, false, ResponseMessage::NO_ITEM);
        }

        if ($news->image && str_contains($news->image, '/storage/')) {
            AttachmentHelper::deleteMedia(config('path.NEWS_PATH'), basename(parse_url($news->image, PHP_URL_PATH)));
        }

        $news->delete();

        return ResponseHelper::responseMessageRequest(StatusCode::OK, true, ResponseMessage::SUCCESS_DELETE);
    }
}
