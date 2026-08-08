-- +goose Up
CREATE TABLE IF NOT EXISTS goals(
    id UUID PRIMARY KEY,
    target INTEGER NOT NULL,
    deadline TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    user_id UUID NOT NULL,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS deposits(
    id UUID PRIMARY KEY,
    amount INTEGER CHECK (amount > 0),
    note TEXT,
    created_at TIMESTAMP NOT NULL,
    goal_id UUID NOT NULL,
    user_id UUID NOT NULL,

    FOREIGN KEY (goal_id) REFERENCES goals(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE deposits;

DROP TABLE goals;

