<?php
namespace App\Http\Controllers;

use App\Enums\ResponseMessage;
use App\Enums\StatusCode;
use App\Helpers\AttachmentHelper;
use App\Helpers\ResponseHelper;
use App\Models\Gallery;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Storage;
use Illuminate\Support\Facades\Validator;
use Illuminate\Support\Str;

class GalleryController extends Controller
{
    /**
     * Display a listing of the resource.
     */
    public function index(Request $request)
    {
        $page   = $request->input('page', 1);
        $limit  = $request->input('limit', 10);
        $skip   = ($page - 1) * $limit;
        $userId = $request->input('user_id'); // Get user_id from request

        $query = Gallery::with('user'); // Start building the query

        if ($userId) {
            $query->where('user_id', $userId); // Apply filter if user_id is present
        }

        $total = $query->count(); // Get total count for pagination

        $data = $query->orderBy('updated_at', 'DESC')
            ->skip($skip)
            ->take($limit)
            ->get();

        return ResponseHelper::responseSuccessDataWithPaginate(
            $data,
            $page,
            $limit,
            $total,
            StatusCode::OK
        );
    }

    /**
     * Store a newly created resource in storage.
     */
    public function store(Request $request)
    {
        $user = Auth::user();

        if (! $user) {
            return ResponseHelper::responseMessageRequest(
                StatusCode::UNAUTHORIZED,
                false,
                ResponseMessage::UNAUTHORIZED
            );
        }

        $validator = Validator::make($request->all(), [
            'image'     => 'required|file|mimes:jpeg,png,jpg,gif,svg|max:2024',
            'caption'   => 'nullable|string',
            'latitude'  => 'nullable|numeric|between:-90,90',
            'longitude' => 'nullable|numeric|between:-180,180',
        ]);

        if ($validator->fails()) {
            return ResponseHelper::responseMessageRequestWithErrors(
                StatusCode::BAD_REQUEST,
                false,
                ResponseMessage::FAILED_CREATE,
                $validator->errors()
            );
        }

        // ================================
        // SAVE IMAGE & GENERATE FULL URL
        // ================================
        $filename = AttachmentHelper::saveImageMedia(
            $request,
            'image',
            config('path.GALLERY_PATH')
        );

        $imageUrl = url(
            Storage::url(
                trim(config('path.GALLERY_PATH'), '/') . '/' . $filename
            )
        );

        // ================================
        // SAVE TO DATABASE (FULL URL)
        // ================================
        $gallery            = new Gallery;
        $gallery->image     = $imageUrl;
        $gallery->caption   = $request->caption;
        $gallery->latitude  = $request->latitude;
        $gallery->longitude = $request->longitude;
        $gallery->slug      = $request->caption ? Str::slug($request->caption) : Str::random(10);
        $gallery->user_id   = $user->id;
        $gallery->save();

        return ResponseHelper::responseMessageRequestWithData(
            StatusCode::CREATED,
            true,
            ResponseMessage::SUCCESS_CREATE,
            $gallery
        );
    }

    /**
     * Display the specified resource.
     */
    public function show($id)
    {
        $gallery = Gallery::with('user')->find($id);
        return ResponseHelper::responseSuccessSingleData(
            $gallery,
            StatusCode::OK
        );
    }

    /**
     * Update the specified resource in storage.
     */
    public function update(Request $request, $id)
    {
        $gallery = Gallery::find($id);
        if (! $gallery) {
            return ResponseHelper::responseMessageRequest(
                StatusCode::NOT_FOUND,
                false,
                ResponseMessage::NOT_ITEM
            );
        }

        $validator = Validator::make($request->all(), [
            'image'     => 'nullable|file|mimes:jpeg,png,jpg,gif,svg|max:2024',
            'caption'   => 'nullable|string',
            'latitude'  => 'nullable|numeric|between:-90,90',
            'longitude' => 'nullable|numeric|between:-180,180',
        ]);

        if ($validator->fails()) {
            return ResponseHelper::responseMessageRequestWithErrors(
                StatusCode::BAD_REQUEST,
                false,
                ResponseMessage::FAILED_UPDATE,
                $validator->errors()
            );
        }

        if ($request->hasFile('image')) {
            // Optional: Delete old image if a new one is uploaded
            if (! empty($gallery->image)) {
                $oldImagePath = Str::after($gallery->image, url('/storage/')); // Extract path from full URL
                AttachmentHelper::deleteMedia(config('path.GALLERY_PATH'), Str::afterLast($oldImagePath, '/'));
            }

            $filename = AttachmentHelper::saveImageMedia(
                $request,
                'image',
                config('path.GALLERY_PATH')
            );

            $gallery->image = url(
                Storage::url(
                    trim(config('path.GALLERY_PATH'), '/') . '/' . $filename
                )
            );
        } elseif ($request->input('remove_existing_image') === 'true') {
            // User explicitly requested to remove existing image
            // Delete old file from storage
            if (! empty($gallery->image)) {
                $oldImagePath = Str::after($gallery->image, url('/storage/')); // Extract path from full URL
                AttachmentHelper::deleteMedia(config('path.GALLERY_PATH'), Str::afterLast($oldImagePath, '/'));
            }
            $gallery->image = null; // Set image to null in database
        }

        $gallery->caption = $request->caption;
        if ($request->has('latitude')) {
            $gallery->latitude = $request->latitude;
        }
        if ($request->has('longitude')) {
            $gallery->longitude = $request->longitude;
        }
        $gallery->slug = $request->caption ? Str::slug($request->caption) : Str::random(10);
        $gallery->save();

        return ResponseHelper::responseMessageRequestWithData(
            StatusCode::OK,
            true,
            ResponseMessage::SUCCESS_UPDATE,
            $gallery
        );
    }

    /**
     * Remove the specified resource from storage.
     */
    public function destroy($id) // Change signature to accept $id
    {
        Log::info('Destroy method called.');
        Log::info('Route parameter ID received: ' . $id); // Log the ID from the route parameter

        $gallery = Gallery::find($id); // Manually find the gallery

        if (! $gallery) { // Check if the gallery was found
            Log::warning('Gallery model not found for ID: ' . $id);
            return ResponseHelper::responseMessageRequest(
                StatusCode::NOT_FOUND,
                false,
                ResponseMessage::NOT_ITEM
            );
        }

        Log::info('Attempting to delete gallery with ID: ' . $gallery->id);

        // OPTIONAL: delete file physically if needed

        $gallery->delete();

        Log::info('Gallery with ID: ' . $gallery->id . ' deleted successfully.');

        return ResponseHelper::responseMessageRequest(
            StatusCode::OK,
            true,
            ResponseMessage::SUCCESS_DELETE
        );
    }

    /**
     * Display an image from storage.
     */
    public function showImage($path)
    {
        if (! Storage::disk('public')->exists($path)) {
            abort(404);
        }

        $file = Storage::disk('public')->get($path);
        $type = Storage::disk('public')->mimeType($path);

        return response($file, 200)->header('Content-Type', $type);
    }
}
