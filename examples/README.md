# Тестирование сценария оформления заказа

1. Убедиться, что по order_id в базе warehouse нет активных reservations
2. Создать в payments заказ в статусе PENDING с данными из `start_process_order_payload.json`
3. Выполнить `start_process_order.sh`