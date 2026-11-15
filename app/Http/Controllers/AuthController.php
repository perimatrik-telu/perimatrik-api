<?php
namespace App\Http\Controllers;

use App\Enums\ErrorCode;
use App\Enums\ResponseMessage;
use App\Enums\StatusCode;
use App\Helpers\ResponseHelper;
use App\Models\Member;
use App\Models\Setting;
use App\Models\User;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;
use Illuminate\Support\Facades\Validator;
use Illuminate\Validation\ValidationException;

class AuthController extends Controller
{
    // User Registration
    public function register(Request $request)
    {
        $registrationEnabled = filter_var(
            optional(Setting::where('key', 'allow_member_registration')->first())->value ?? '1',
            FILTER_VALIDATE_BOOLEAN
        );

        if (! $registrationEnabled) {
            throw ValidationException::withMessages([
                'register' => ['Registrasi anggota sedang dinonaktifkan oleh admin.'],
            ]);
        }

        $validated = $request->validate([
            'no_anggota' => 'required|string|max:255',
            'nama_anggota' => 'required|string|max:255',
            'email'    => 'required|string|email|max:255|unique:users',
            'password' => 'required|string|min:8|confirmed',
        ]);

        $member = Member::where('no_anggota', $validated['no_anggota'])->first();

        if (! $member) {
            throw ValidationException::withMessages([
                'no_anggota' => ['Nomor anggota tidak ditemukan.'],
            ]);
        }

        if (Str::lower(trim($member->nama_lengkap)) !== Str::lower(trim($validated['nama_anggota']))) {
            throw ValidationException::withMessages([
                'nama_anggota' => ['Nama anggota tidak sesuai dengan nomor anggota.'],
            ]);
        }

        if ($member->user_id) {
            throw ValidationException::withMessages([
                'no_anggota' => ['Nomor anggota ini sudah terhubung ke akun pengguna.'],
            ]);
        }

        $user = DB::transaction(function () use ($validated, $member) {
            $user = User::create([
                'name'     => $member->nama_lengkap,
                'email'    => $validated['email'],
                'password' => Hash::make($validated['password']),
                'role'     => 'member',
            ]);

            $member->update([
                'user_id' => $user->id,
            ]);

            return $user;
        });

        return ResponseHelper::responseMessageRequestWithData(
            StatusCode::CREATED,
            true,
            'Registrasi berhasil. Silakan login menggunakan email dan password Anda.',
            $user
        );
    }

    // User Login
    public function login(Request $request)
    {
        $request->validate([
            'email'    => 'required|email',
            'password' => 'required',
        ]);

        $user = User::where('email', $request->email)->first();

        if (! $user || ! Hash::check($request->password, $user->password)) {
            return ResponseHelper::error(
                StatusCode::UNPROCESSABLE_ENTITY,
                'The provided credentials are incorrect.',
                ['email' => ['The provided credentials are incorrect.']],
                ErrorCode::INVALID_CREDENTIALS
            );
        }

        $token = $user->createToken('auth_token')->plainTextToken;

        return ResponseHelper::success(StatusCode::OK, ResponseMessage::LOGIN_SUCCESS, [
            'access_token' => $token,
            'token_type'   => 'Bearer',
        ]);
    }

    // Get Authenticated User
    public function currentUser(Request $request)
    {
        return ResponseHelper::responseMessageRequestWithData(StatusCode::OK, true, ResponseMessage::FOUND, $request->user());
    }

    // User Logout
    public function logout(Request $request)
    {
        $request->user()->currentAccessToken()->delete();

        return ResponseHelper::success(StatusCode::OK, ResponseMessage::LOGOUT_SUCCESS);
    }

    /**
     * Change the user's password.
     */
    public function changePassword(Request $request)
    {
        $validator = Validator::make($request->all(), [
            'old_password' => 'required|string',
            'password' => 'required|string|min:8|confirmed',
        ]);

        if ($validator->fails()) {
            return ResponseHelper::responseMessageRequestWithErrors(
                StatusCode::UNPROCESSABLE_ENTITY,
                false,
                'Validasi gagal.',
                $validator->errors()
            );
        }

        $user = auth()->user();

        if (!Hash::check($request->old_password, $user->password)) {
            return ResponseHelper::responseMessageRequest(
                StatusCode::BAD_REQUEST,
                false,
                'Password lama tidak sesuai.'
            );
        }

        $user->password = Hash::make($request->password);
        $user->save();

        return ResponseHelper::responseMessageRequest(
            StatusCode::OK,
            true,
            'Password berhasil diubah.'
        );
    }


    /**
     * Verify the user's old password.
     */
    public function verifyPassword(Request $request)
    {
        $validator = Validator::make($request->all(), [
            'old_password' => 'required|string',
        ]);

        if ($validator->fails()) {
            return ResponseHelper::responseMessageRequestWithErrors(
                StatusCode::UNPROCESSABLE_ENTITY,
                false,
                'Validasi gagal.',
                $validator->errors()
            );
        }

        $user = auth()->user();

        if (!Hash::check($request->old_password, $user->password)) {
            return ResponseHelper::responseMessageRequest(
                StatusCode::BAD_REQUEST,
                false,
                'Password lama tidak sesuai.'
            );
        }

        return ResponseHelper::responseMessageRequest(
            StatusCode::OK,
            true,
            'Password lama sesuai.'
        );
    }

}

