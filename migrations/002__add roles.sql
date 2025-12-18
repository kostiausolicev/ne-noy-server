-- +goose Up
INSERT INTO public.roles (id, name, display_name)
VALUES ('9ef01b95-6d87-4115-80df-7085a647bf36', 'default', 'Внешний участник')
ON CONFLICT (id) DO NOTHING;

INSERT INTO public.roles (id, name, display_name)
VALUES ('704a0251-f27d-4f54-b09b-e17f8be2d905', 'admin', 'Администратор')
ON CONFLICT (id) DO NOTHING;

INSERT INTO public.roles (id, name, display_name)
VALUES ('1aae97ca-31f7-4e8f-8f0a-885131a98e47', 'hike', 'Участник тур клуба')
ON CONFLICT (id) DO NOTHING;

-- +goose Down
DELETE
FROM public.roles
WHERE id IN (
             '9ef01b95-6d87-4115-80df-7085a647bf36',
             '704a0251-f27d-4f54-b09b-e17f8be2d905',
             '1aae97ca-31f7-4e8f-8f0a-885131a98e47'
    );