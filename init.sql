CREATE TABLE products(
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name text NOT NULL,
    description text NOT NULL,
    product_search_vector tsvector, -- combined `name` and `description` fields for easy search
    price numeric(8, 2)
);

CREATE INDEX product_search_idx ON products USING gin(product_search_vector);

-- 1. Create the Trigger Function
-- This function uses the 'english' configuration (dictionary) to process the text
-- in the `description` column and assign the resulting tsvector to the
-- `product_search_vector` column.
CREATE OR REPLACE FUNCTION update_product_search_vector()
    RETURNS TRIGGER
    AS $$
BEGIN
    NEW.product_search_vector := to_tsvector('english', NEW.description);
    RETURN NEW;
END;
$$
LANGUAGE plpgsql;

-- 2. Create the Trigger
-- This trigger executes the function before an INSERT or an UPDATE happens
-- on the products table, ensuring the tsvector is always fresh.
CREATE TRIGGER product_tsvector_update
    BEFORE INSERT OR UPDATE OF description ON products
    FOR EACH ROW
    EXECUTE FUNCTION update_product_search_vector();

