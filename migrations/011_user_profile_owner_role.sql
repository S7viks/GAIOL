-- Personal-tenant signups should be owners so they can manage provider keys.
-- Run in Supabase SQL Editor after prior migrations.

CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO public.user_profiles (id, email, tenant_id, role)
    VALUES (
        NEW.id,
        NEW.email,
        NEW.id,
        'owner'
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Upgrade existing solo tenants created with default role=user.
UPDATE public.user_profiles
SET role = 'owner', updated_at = NOW()
WHERE role = 'user'
  AND (tenant_id = id OR tenant_id IS NULL);
