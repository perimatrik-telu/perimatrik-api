<?php

namespace App\Http\Controllers;

use App\Models\Comment;
use App\Enums\StatusCode;
use App\Enums\ResponseMessage;
use App\Helpers\ResponseHelper;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Validator;

class CommentController extends Controller
{
    /**
     * Get comments for a specific article or news.
     */
    public function index(Request $request)
    {
        $validator = Validator::make($request->all(), [
            'article_id' => 'nullable|uuid|exists:articles,id',
            'news_id'    => 'nullable|uuid|exists:news,id',
        ]);

        if ($validator->fails()) {
            return ResponseHelper::responseMessageRequest(StatusCode::UNPROCESSABLE_ENTITY, false, "Invalid ID provided.");
        }

        $query = Comment::with('user', 'replies')
            ->whereNull('parent_id') // Only get top-level comments
            ->orderBy('created_at', 'desc');

        if ($request->filled('article_id')) {
            $query->where('article_id', $request->article_id);
        } elseif ($request->filled('news_id')) {
            $query->where('news_id', $request->news_id);
        } else {
            return ResponseHelper::responseMessageRequest(StatusCode::BAD_REQUEST, false, "Context ID is required.");
        }

        $comments = $query->get();
        return ResponseHelper::responseSuccessAllData($comments, StatusCode::OK);
    }

    /**
     * Store a new comment or reply.
     */
    public function store(Request $request)
    {
        $validator = Validator::make($request->all(), [
            'body'       => 'required|string',
            'article_id' => 'nullable|uuid|exists:articles,id',
            'news_id'    => 'nullable|uuid|exists:news,id',
            'parent_id'  => 'nullable|uuid|exists:comments,id',
        ]);

        if ($validator->fails()) {
            return ResponseHelper::responseMessageRequestWithErrors(
                StatusCode::UNPROCESSABLE_ENTITY,
                false,
                "Validation failed",
                $validator->errors()
            );
        }

        $comment = Comment::create([
            'body'       => $request->body,
            'user_id'    => auth()->id(),
            'article_id' => $request->article_id,
            'news_id'    => $request->news_id,
            'parent_id'  => $request->parent_id,
        ]);

        return ResponseHelper::responseSuccessSingleData($comment->load('user'), StatusCode::CREATED);
    }

    /**
     * Remove a comment.
     */
    public function destroy(string $id)
    {
        $comment = Comment::find($id);

        if (!$comment) {
            return ResponseHelper::responseMessageRequest(StatusCode::NOT_FOUND, false, ResponseMessage::NO_ITEM);
        }

        // Only owner or admin can delete
        if (auth()->user()->role !== 'admin' && auth()->id() !== $comment->user_id) {
            return ResponseHelper::responseMessageRequest(StatusCode::FORBIDDEN, false, "Unauthorized to delete this comment.");
        }

        $comment->delete();

        return ResponseHelper::responseMessageRequest(StatusCode::OK, true, "Comment deleted successfully.");
    }
}
