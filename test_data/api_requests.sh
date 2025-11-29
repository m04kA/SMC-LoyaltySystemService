#!/bin/bash

# ==========================================
# LoyaltySystemService API Test Commands
# ==========================================
# Набор curl команд для тестирования всех endpoints
#
# Использование:
#   chmod +x test_data/api_requests.sh
#   ./test_data/api_requests.sh
#
# Или вызывать отдельные функции:
#   source test_data/api_requests.sh
#   tc_1_1  # Запустить конкретный тест
#
# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Конфигурация
BASE_URL="http://localhost:8084"

# Функции для вывода
print_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

print_test() {
    echo -e "${YELLOW}[TEST] $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# ==========================================
# 1. Get Loyalty Card (Получение карты лояльности)
# ==========================================

# TC-1.1: Получение существующей карты лояльности
tc_1_1() {
    print_test "TC-1.1: Получение существующей карты лояльности (user 123456789, company 1)"
    curl -s -X GET "${BASE_URL}/api/v1/loyalty-cards?userId=123456789&companyId=1" \
      -H "Content-Type: application/json" | jq .
}

# TC-1.2: Получение карты для несуществующей компании
tc_1_2() {
    print_test "TC-1.2: Несуществующая компания (ожидается 404)"
    curl -s -X GET "${BASE_URL}/api/v1/loyalty-cards?userId=123456789&companyId=99999" \
      -H "Content-Type: application/json" -w "\nHTTP Code: %{http_code}\n"
}

# TC-1.3: Карта не найдена
tc_1_3() {
    print_test "TC-1.3: Карта не найдена для пользователя (ожидается 404)"
    curl -s -X GET "${BASE_URL}/api/v1/loyalty-cards?userId=999999999&companyId=1" \
      -H "Content-Type: application/json" -w "\nHTTP Code: %{http_code}\n"
}

# TC-1.4: Конфигурация лояльности отключена
tc_1_4() {
    print_test "TC-1.4: Программа лояльности отключена (ожидается 404)"
    curl -s -X GET "${BASE_URL}/api/v1/loyalty-cards?userId=123456789&companyId=2" \
      -H "Content-Type: application/json" -w "\nHTTP Code: %{http_code}\n"
}

# TC-1.5: Невалидные параметры
tc_1_5() {
    print_test "TC-1.5: Невалидные параметры - отсутствует userId (ожидается 400)"
    curl -s -X GET "${BASE_URL}/api/v1/loyalty-cards?companyId=1" \
      -H "Content-Type: application/json" -w "\nHTTP Code: %{http_code}\n"
}

# TC-1.6: Невалидные параметры - отсутствует companyId
tc_1_6() {
    print_test "TC-1.6: Невалидные параметры - отсутствует companyId (ожидается 400)"
    curl -s -X GET "${BASE_URL}/api/v1/loyalty-cards?userId=123456789" \
      -H "Content-Type: application/json" -w "\nHTTP Code: %{http_code}\n"
}

# ==========================================
# 2. Create Loyalty Card (Создание карты лояльности)
# ==========================================

# TC-2.1: Успешное создание карты лояльности
tc_2_1() {
    print_test "TC-2.1: Успешное создание карты лояльности (ожидается 201)"
    curl -s -X POST "${BASE_URL}/api/v1/loyalty-cards" \
      -H "Content-Type: application/json" \
      -d '{
        "userId": 987654321,
        "companyId": 1
      }' -w "\nHTTP Code: %{http_code}\n" | jq .
}

# TC-2.2: Попытка создать дубликат карты
tc_2_2() {
    print_test "TC-2.2: Попытка создать дубликат карты (ожидается 409)"
    curl -s -X POST "${BASE_URL}/api/v1/loyalty-cards" \
      -H "Content-Type: application/json" \
      -d '{
        "userId": 123456789,
        "companyId": 1
      }' -w "\nHTTP Code: %{http_code}\n"
}

# TC-2.3: Создание карты при отключенной конфигурации
tc_2_3() {
    print_test "TC-2.3: Программа лояльности отключена (ожидается 404)"
    curl -s -X POST "${BASE_URL}/api/v1/loyalty-cards" \
      -H "Content-Type: application/json" \
      -d '{
        "userId": 123456789,
        "companyId": 2
      }' -w "\nHTTP Code: %{http_code}\n"
}

# TC-2.4: Несуществующая компания
tc_2_4() {
    print_test "TC-2.4: Несуществующая компания (ожидается 404)"
    curl -s -X POST "${BASE_URL}/api/v1/loyalty-cards" \
      -H "Content-Type: application/json" \
      -d '{
        "userId": 123456789,
        "companyId": 99999
      }' -w "\nHTTP Code: %{http_code}\n"
}

# TC-2.5: Невалидные данные - отсутствует userId
tc_2_5() {
    print_test "TC-2.5: Невалидные данные - отсутствует userId (ожидается 400)"
    curl -s -X POST "${BASE_URL}/api/v1/loyalty-cards" \
      -H "Content-Type: application/json" \
      -d '{
        "companyId": 1
      }' -w "\nHTTP Code: %{http_code}\n"
}

# TC-2.6: Невалидные данные - отсутствует companyId
tc_2_6() {
    print_test "TC-2.6: Невалидные данные - отсутствует companyId (ожидается 400)"
    curl -s -X POST "${BASE_URL}/api/v1/loyalty-cards" \
      -H "Content-Type: application/json" \
      -d '{
        "userId": 123456789
      }' -w "\nHTTP Code: %{http_code}\n"
}

# TC-2.7: Невалидный JSON
tc_2_7() {
    print_test "TC-2.7: Невалидный JSON (ожидается 400)"
    curl -s -X POST "${BASE_URL}/api/v1/loyalty-cards" \
      -H "Content-Type: application/json" \
      -d 'invalid json' -w "\nHTTP Code: %{http_code}\n"
}

# ==========================================
# 3. Configure Loyalty (Настройка программы лояльности)
# ==========================================

# TC-3.1: Создание конфигурации менеджером
tc_3_1() {
    print_test "TC-3.1: Создание конфигурации программы лояльности менеджером (ожидается 200)"
    curl -s -X POST "${BASE_URL}/api/v1/companies/1/loyalty-config" \
      -H "Content-Type: application/json" \
      -H "X-User-ID: 777777777" \
      -d '{
        "discountPercentage": 10.5,
        "isEnabled": true
      }' -w "\nHTTP Code: %{http_code}\n" | jq .
}

# TC-3.2: Обновление существующей конфигурации
tc_3_2() {
    print_test "TC-3.2: Обновление конфигурации (изменение скидки)"
    curl -s -X POST "${BASE_URL}/api/v1/companies/1/loyalty-config" \
      -H "Content-Type: application/json" \
      -H "X-User-ID: 777777777" \
      -d '{
        "discountPercentage": 15.0,
        "isEnabled": true
      }' -w "\nHTTP Code: %{http_code}\n" | jq .
}

# TC-3.3: Отключение программы лояльности
tc_3_3() {
    print_test "TC-3.3: Отключение программы лояльности"
    curl -s -X POST "${BASE_URL}/api/v1/companies/1/loyalty-config" \
      -H "Content-Type: application/json" \
      -H "X-User-ID: 777777777" \
      -d '{
        "discountPercentage": 10.0,
        "isEnabled": false
      }' -w "\nHTTP Code: %{http_code}\n" | jq .
}

# TC-3.4: Включение программы лояльности обратно
tc_3_4() {
    print_test "TC-3.4: Включение программы лояльности обратно"
    curl -s -X POST "${BASE_URL}/api/v1/companies/1/loyalty-config" \
      -H "Content-Type: application/json" \
      -H "X-User-ID: 777777777" \
      -d '{
        "discountPercentage": 10.0,
        "isEnabled": true
      }' -w "\nHTTP Code: %{http_code}\n" | jq .
}

# TC-3.5: Попытка настройки не-менеджером
tc_3_5() {
    print_test "TC-3.5: Попытка настройки не-менеджером (ожидается 403)"
    curl -s -X POST "${BASE_URL}/api/v1/companies/1/loyalty-config" \
      -H "Content-Type: application/json" \
      -H "X-User-ID: 123456789" \
      -d '{
        "discountPercentage": 50.0,
        "isEnabled": true
      }' -w "\nHTTP Code: %{http_code}\n"
}

# TC-3.6: Несуществующая компания
tc_3_6() {
    print_test "TC-3.6: Несуществующая компания (ожидается 404)"
    curl -s -X POST "${BASE_URL}/api/v1/companies/99999/loyalty-config" \
      -H "Content-Type: application/json" \
      -H "X-User-ID: 777777777" \
      -d '{
        "discountPercentage": 10.0,
        "isEnabled": true
      }' -w "\nHTTP Code: %{http_code}\n"
}

# TC-3.7: Невалидная скидка (> 100)
tc_3_7() {
    print_test "TC-3.7: Невалидная скидка > 100% (ожидается 400)"
    curl -s -X POST "${BASE_URL}/api/v1/companies/1/loyalty-config" \
      -H "Content-Type: application/json" \
      -H "X-User-ID: 777777777" \
      -d '{
        "discountPercentage": 150.0,
        "isEnabled": true
      }' -w "\nHTTP Code: %{http_code}\n"
}

# TC-3.8: Невалидная скидка (< 0)
tc_3_8() {
    print_test "TC-3.8: Невалидная скидка < 0% (ожидается 400)"
    curl -s -X POST "${BASE_URL}/api/v1/companies/1/loyalty-config" \
      -H "Content-Type: application/json" \
      -H "X-User-ID: 777777777" \
      -d '{
        "discountPercentage": -10.0,
        "isEnabled": true
      }' -w "\nHTTP Code: %{http_code}\n"
}

# TC-3.9: Без X-User-ID
tc_3_9() {
    print_test "TC-3.9: Запрос без X-User-ID (ожидается 401)"
    curl -s -X POST "${BASE_URL}/api/v1/companies/1/loyalty-config" \
      -H "Content-Type: application/json" \
      -d '{
        "discountPercentage": 10.0,
        "isEnabled": true
      }' -w "\nHTTP Code: %{http_code}\n"
}

# ==========================================
# Prometheus Metrics
# ==========================================

tc_metrics() {
    print_test "Проверка Prometheus метрик"
    curl -s -X GET "${BASE_URL}/metrics" | grep -E "^(http_|db_)" | head -20
}

# ==========================================
# Главное меню
# ==========================================

show_menu() {
    echo ""
    print_header "LoyaltySystemService API Test Suite"
    echo "Выберите группу тестов:"
    echo "  1) Get Loyalty Card (TC-1.*)"
    echo "  2) Create Loyalty Card (TC-2.*)"
    echo "  3) Configure Loyalty (TC-3.*)"
    echo "  4) Prometheus Metrics"
    echo "  s) Smoke Tests (основные сценарии)"
    echo "  a) All Tests (все тесты)"
    echo "  q) Quit"
    echo ""
}

# Smoke тесты
run_smoke_tests() {
    print_header "SMOKE TESTS"
    tc_1_1
    sleep 1
    tc_2_1
    sleep 1
    tc_3_1
    sleep 1
    tc_metrics
    print_success "Smoke tests completed"
}

# Запуск всех тестов группы 1
run_group_1() {
    print_header "GROUP 1: Get Loyalty Card Tests"
    tc_1_1; sleep 0.5
    tc_1_2; sleep 0.5
    tc_1_3; sleep 0.5
    tc_1_4; sleep 0.5
    tc_1_5; sleep 0.5
    tc_1_6
    print_success "Group 1 completed"
}

# Запуск всех тестов группы 2
run_group_2() {
    print_header "GROUP 2: Create Loyalty Card Tests"
    tc_2_1; sleep 0.5
    tc_2_2; sleep 0.5
    tc_2_3; sleep 0.5
    tc_2_4; sleep 0.5
    tc_2_5; sleep 0.5
    tc_2_6; sleep 0.5
    tc_2_7
    print_success "Group 2 completed"
}

# Запуск всех тестов группы 3
run_group_3() {
    print_header "GROUP 3: Configure Loyalty Tests"
    tc_3_1; sleep 0.5
    tc_3_2; sleep 0.5
    tc_3_3; sleep 0.5
    tc_3_4; sleep 0.5
    tc_3_5; sleep 0.5
    tc_3_6; sleep 0.5
    tc_3_7; sleep 0.5
    tc_3_8; sleep 0.5
    tc_3_9
    print_success "Group 3 completed"
}

# Интерактивный режим
interactive_mode() {
    while true; do
        show_menu
        read -p "Ваш выбор: " choice
        case $choice in
            1) run_group_1 ;;
            2) run_group_2 ;;
            3) run_group_3 ;;
            4) tc_metrics ;;
            s|S) run_smoke_tests ;;
            a|A) echo "Запуск всех тестов..."
                 run_smoke_tests; sleep 2
                 run_group_1; sleep 2
                 run_group_2; sleep 2
                 run_group_3; sleep 2
                 print_success "All tests completed" ;;
            q|Q) echo "Выход..."; exit 0 ;;
            *) print_error "Неверный выбор" ;;
        esac
    done
}

# Если скрипт запущен напрямую (не через source)
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    # Проверка зависимостей
    if ! command -v jq > /dev/null 2>&1; then
        print_error "jq не установлен. Установите: brew install jq"
        exit 1
    fi

    # Проверка доступности сервиса
    if ! curl -s -f "${BASE_URL}/metrics" > /dev/null 2>&1; then
        print_error "LoyaltySystemService недоступен на ${BASE_URL}"
        print_error "Запустите сервис: make docker-up"
        exit 1
    fi

    # Запуск интерактивного режима
    interactive_mode
fi
