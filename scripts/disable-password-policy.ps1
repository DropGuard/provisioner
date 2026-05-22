# scripts/disable-password-policy.ps1
# Disables Windows Server password complexity and minimum password length policies in CI.
secedit /export /cfg sec.cfg
(Get-Content sec.cfg) -replace 'PasswordComplexity = 1', 'PasswordComplexity = 0' -replace 'MinimumPasswordLength = \d+', 'MinimumPasswordLength = 0' | Set-Content sec.cfg
secedit /configure /db secedit.sdb /cfg sec.cfg /areas SECURITYPOLICY
Remove-Item sec.cfg -Force -ErrorAction SilentlyContinue
