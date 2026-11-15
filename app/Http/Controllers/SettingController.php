<?php

namespace App\Http\Controllers;

use App\Models\Setting;
use App\Enums\StatusCode;
use App\Helpers\ResponseHelper;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Validator;

class SettingController extends Controller
{
    private function getSettingValue(string $key, mixed $default = '')
    {
        $setting = Setting::where('key', $key)->first();
        return $setting ? $setting->value : $default;
    }

    /**
     * Get all contact information settings.
     */
    public function getContactInfo()
    {
        $data = [
            'contact_email' => $this->getSettingValue('contact_email', ''),
            'contact_phone' => $this->getSettingValue('contact_phone', ''),
            'contact_address' => $this->getSettingValue('contact_address', ''),
            'contact_maps' => $this->getSettingValue('contact_maps', ''),
            'allow_member_registration' => filter_var(
                $this->getSettingValue('allow_member_registration', '1'),
                FILTER_VALIDATE_BOOLEAN
            ),
        ];

        return ResponseHelper::responseSuccessSingleData($data, StatusCode::OK);
    }

    /**
     * Update contact information settings (Admin only).
     */
    public function updateContactInfo(Request $request)
    {
        $validator = Validator::make($request->all(), [
            'contact_email'   => 'nullable|email|max:255',
            'contact_phone'   => 'nullable|string|max:255',
            'contact_address' => 'nullable|string',
            'contact_maps'    => 'nullable|string',
            'allow_member_registration' => 'required|boolean',
        ]);

        if ($validator->fails()) {
            return ResponseHelper::responseMessageRequestWithErrors(
                StatusCode::UNPROCESSABLE_ENTITY,
                false,
                "Validation failed",
                $validator->errors()
            );
        }

        foreach ($request->only([
            'contact_email',
            'contact_phone',
            'contact_address',
            'contact_maps',
            'allow_member_registration',
        ]) as $key => $value) {
            Setting::updateOrCreate(
                ['key' => $key],
                ['value' => is_bool($value) ? ($value ? '1' : '0') : $value]
            );
        }

        return ResponseHelper::responseMessageRequest(StatusCode::OK, true, "Contact information updated successfully.");
    }
}
