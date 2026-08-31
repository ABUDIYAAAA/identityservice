package mailer

const InviteEmailHTML = `
<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <h2>You're Invited!</h2>
    <p>You have been invited to join the platform as <strong>{{.Role}}</strong>.</p>
    <p>Click the link below to set up your password and complete your registration:</p>
    <p>
        <a href="{{.ActionURL}}" style="background-color: #007bff; color: white; padding: 10px 20px; text-decoration: none; border-radius: 4px; display: inline-block;">Accept Invitation</a>
    </p>
    <p><small>This invitation will expire in {{.ExpiresIn}}.</small></p>
</body>
</html>
`

const ResetPasswordHTML = `
<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <h2>Password Reset Request</h2>
    <p>We received a request to reset your password. Click the link below to set a new one:</p>
    <p>
        <a href="{{.ActionURL}}" style="background-color: #dc3545; color: white; padding: 10px 20px; text-decoration: none; border-radius: 4px; display: inline-block;">Reset Password</a>
    </p>
    <p><small>If you did not request this, you can safely ignore this email. Link expires in {{.ExpiresIn}}.</small></p>
</body>
</html>
`
