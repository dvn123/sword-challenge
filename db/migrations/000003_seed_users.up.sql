INSERT INTO roles (name)
VALUES ('technician')
ON DUPLICATE KEY UPDATE name=name;
INSERT INTO roles (name)
VALUES ('manager')
ON DUPLICATE KEY UPDATE name=name;

SET @admin_role_id = (SELECT id
                      FROM roles
                      WHERE name = 'manager'
                      LIMIT 1);
SET @user_role_id = (SELECT id
                     FROM roles
                     WHERE name = 'technician'
                     LIMIT 1);

INSERT INTO users (username, role_id)
VALUES ('joel', @user_role_id);

INSERT INTO users (username, role_id)
VALUES ('dvn', @admin_role_id);