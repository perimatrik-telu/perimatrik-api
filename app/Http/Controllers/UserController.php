<?php

namespace App\Http\Controllers;

use App\Models\User;
use App\Models\Member;
use App\Enums\StatusCode;
use App\Enums\ResponseMessage;
use App\Helpers\ResponseHelper;
use App\Helpers\AttachmentHelper;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Facades\Validator;
use Illuminate\Validation\Rule;

class UserController extends Controller
{
    /**
     * Display a listing of the users.
     */
    public function index()
    {
        $users = User::with('member.angkatan')->orderBy('created_at', 'desc')->get();
        return ResponseHelper::responseSuccessAllData($users, StatusCode::OK);
    }

    /**
     * Store a newly created user in storage.
     */
    public function store(Request $request)
    {
        $validator = Validator::make($request->all(), [
            'name'     => 'required|string|max:255',
            'email'    => 'required|string|email|max:255|unique:users',
            'password' => 'required|string|min:8',
            'role'     => ['required', Rule::in(['admin', 'officer', 'member'])],
            'avatar'   => 'nullable|file|mimes:jpeg,png,jpg,gif,svg|max:2048',
        ]);

        if ($validator->fails()) {
            return ResponseHelper::responseMessageRequestWithErrors(
                StatusCode::UNPROCESSABLE_ENTITY,
                false,
                ResponseMessage::FAILED_CREATE,
                $validator->errors()
            );
        }

        $avatarUrl = null;
        if ($request->hasFile('avatar')) {
            $filename  = AttachmentHelper::saveImageMedia($request, 'avatar', 'avatars');
            $avatarUrl = config('app.url') . '/storage/avatars/' . $filename;
        }

        $user = User::create([
            'name'     => $request->name,
            'email'    => $request->email,
            'password' => Hash::make($request->password),
            'role'     => $request->role,
            'avatar'   => $avatarUrl,
        ]);

        // Link to existing member if provided
        if ($request->filled('member_id')) {
            $member = Member::find($request->member_id);
            if ($member) {
                $member->update([
                    'user_id'      => $user->id,
                    'nama_lengkap' => $user->name,
                    'avatar'       => $user->avatar,
                ]);
            }
        }

        return ResponseHelper::responseSuccessSingleData($user, StatusCode::CREATED);
    }

    /**
     * Display the specified user.
     */
    public function show(string $id)
    {
        // Security: only allow own profile or admin
        if (auth()->user()->role !== 'admin' && auth()->id() !== $id) {
            return ResponseHelper::responseMessageRequest(StatusCode::FORBIDDEN, false, "Forbidden. You can only view your own profile.");
        }

        $user = User::with('member.angkatan')->find($id);

        if (!$user) {
            return ResponseHelper::responseMessageRequest(
                StatusCode::NOT_FOUND,
                false,
                ResponseMessage::NO_ITEM
            );
        }

        return ResponseHelper::responseSuccessSingleData($user, StatusCode::OK);
    }

    /**
     * Update the specified user in storage.
     */
    public function update(Request $request, string $id)
    {
        // Security: only allow own profile or admin
        if (auth()->user()->role !== 'admin' && auth()->id() !== $id) {
            return ResponseHelper::responseMessageRequest(StatusCode::FORBIDDEN, false, "Forbidden. You can only update your own profile.");
        }

        $user = User::find($id);

        if (!$user) {
            return ResponseHelper::responseMessageRequest(
                StatusCode::NOT_FOUND,
                false,
                ResponseMessage::NO_ITEM
            );
        }

        $rules = [
            'name'     => 'required|string|max:255',
            'email'    => ['required', 'string', 'email', 'max:255', Rule::unique('users')->ignore($user->id)],
            'password' => 'nullable|string|min:8',
            'avatar'   => 'nullable|file|mimes:jpeg,png,jpg,gif,svg|max:2048',
            // Member fields
            'no_anggota'    => 'nullable|string|max:255',
            'nama_lapangan' => 'nullable|string|max:255',
            'domisili'      => 'nullable|string|max:255',
            'kontak'        => 'nullable|string|max:255',
            'instagram'     => 'nullable|string|max:255',
            'facebook'      => 'nullable|string|max:255',
            'linkedin'      => 'nullable|string|max:255',
            'tiktok'        => 'nullable|string|max:255',
            'youtube'       => 'nullable|string|max:255',
        ];

        // Only admin can change roles
        if (auth()->user()->role === 'admin') {
            $rules['role'] = ['required', Rule::in(['admin', 'officer', 'member'])];
        }

        $validator = Validator::make($request->all(), $rules);

        if ($validator->fails()) {
            return ResponseHelper::responseMessageRequestWithErrors(
                StatusCode::UNPROCESSABLE_ENTITY,
                false,
                ResponseMessage::FAILED_UPDATE,
                $validator->errors()
            );
        }

        if ($request->hasFile('avatar')) {
            // Delete old avatar
            if ($user->avatar) {
                $oldFilename = basename(parse_url($user->avatar, PHP_URL_PATH));
                AttachmentHelper::deleteMedia('avatars', $oldFilename);
            }
            $filename     = AttachmentHelper::saveImageMedia($request, 'avatar', 'avatars');
            $user->avatar = config('app.url') . '/storage/avatars/' . $filename;
        }

        $user->name  = $request->name;
        $user->email = $request->email;
        
        // Only update role if admin provides it
        if (auth()->user()->role === 'admin' && $request->filled('role')) {
            $user->role  = $request->role;
        }

        if ($request->filled('password')) {
            $user->password = Hash::make($request->password);
        }

        $user->save();

        // Sync with Member table
        $member = Member::where('user_id', $user->id)->first();
        if ($member) {
            $syncData = [
                'nama_lengkap'  => $user->name,
                'avatar'        => $user->avatar,
            ];

            // Only sync extra fields if provided in request
            if ($request->has('no_anggota')) $syncData['no_anggota'] = $request->no_anggota;
            if ($request->has('nama_lapangan')) $syncData['nama_lapangan'] = $request->nama_lapangan;
            if ($request->has('domisili')) $syncData['domisili'] = $request->domisili;
            if ($request->has('kontak')) $syncData['kontak'] = $request->kontak;
            if ($request->has('instagram')) $syncData['instagram'] = $request->instagram;
            if ($request->has('facebook')) $syncData['facebook'] = $request->facebook;
            if ($request->has('linkedin')) $syncData['linkedin'] = $request->linkedin;
            if ($request->has('tiktok')) $syncData['tiktok'] = $request->tiktok;
            if ($request->has('youtube')) $syncData['youtube'] = $request->youtube;

            $member->update($syncData);
        }

        return ResponseHelper::responseSuccessSingleData($user->load('member.angkatan'), StatusCode::OK);
    }

    /**
     * Remove the specified user from storage.
     */
    public function destroy(string $id)
    {
        $user = User::find($id);

        if (!$user) {
            return ResponseHelper::responseMessageRequest(
                StatusCode::NOT_FOUND,
                false,
                ResponseMessage::NO_ITEM
            );
        }

        // Prevent deleting self
        if ($user->id === auth()->id()) {
            return ResponseHelper::responseMessageRequest(
                StatusCode::BAD_REQUEST,
                false,
                "You cannot delete your own account."
            );
        }

        $user->delete();

        return ResponseHelper::responseMessageRequest(
            StatusCode::OK,
            true,
            "User deleted successfully."
        );
    }
}
