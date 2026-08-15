CREATE OR REPLACE FUNCTION notify_saga_runnable() RETURNS trigger AS $$
BEGIN
    IF NEW.execution_state = 'RUNNABLE'
        AND NEW.status IN ('PENDING','RUNNING','COMPENSATING','VERIFYING') THEN
        -- пустой payload: смысл "есть работа, иди проверь",
        -- сам инстанс выберет TakeBatch с правильным приоритетом
        PERFORM pg_notify('saga_runnable', '');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_saga_runnable
    AFTER INSERT OR UPDATE OF execution_state ON saga_instance
    FOR EACH ROW EXECUTE FUNCTION notify_saga_runnable();