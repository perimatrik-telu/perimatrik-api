<?php
namespace App\Enums;

enum ResponseMessage: string {
    case INTERNAL_SERVER_ERROR = "Internal Server Error";
    case FOUND                 = "Item found";
    case NO_ITEM               = "Item not found";
    /* Create */
    case SUCCESS_CREATE = "Successfully created item";
    case FAILED_CREATE  = "Failed to create item";
    /* Update */
    case SUCCESS_UPDATE = "Successfully updated item";
    case FAILED_UPDATE  = "Failed to update item";
    /* Delete */
    case SUCCESS_DELETE = "Successfully deleted item";
    case FAILED_DELETE  = "Failed to delete item";
    /* Authorized */
    case LOGIN_SUCCESS  = "Login Successfully";
    case LOGIN_FAILED   = "Login Failed";
    case AUTHORIZED     = "Success Authorized";
    case UNAUTHORIZED   = "Unauthorized";
    case LOGOUT_SUCCESS = "Logout Successfully";
    case LOGOUT_FAILED  = "Logout Failed";
    /* Payment */
    case PAYMENT_NOT_FOUND      = "Payment not found";
    case PAYMENT_SUCCESS_UPDATE = "Payment successfully updated";
    case PAYMENT_FAILED_UPDATE  = "Payment failed updated";
    /* Plate */
    case PLATE_NOT_VALID = " Nomor kendaraan Anda tidak terdaftar atau di luar jangkauan provinsi Jawa Barat.";
    /* Change Password */
    case CHANGE_PASSWORD_SUCCESS_UPDATED = "Your password has been successfully updated!";
    case CHANGE_PASSWORD_FAILED_UPDATED  = "Password update failed. Please try again.";
}
