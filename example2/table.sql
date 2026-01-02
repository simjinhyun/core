-- 데이터베이스 생성 (없으면 생성)
CREATE DATABASE IF NOT EXISTS testdb;
USE testdb;

-- 주소록 테이블이 있으면 삭제
DROP TABLE IF EXISTS address_book;

-- 새로 테이블 생성 (created_at은 TIMESTAMP)
CREATE TABLE address_book (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    age INT NOT NULL,
    phone VARCHAR(20),
    address VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 샘플 데이터 삽입
INSERT INTO address_book (name, age, phone, address, created_at) VALUES
('김철수', 32, '010-1234-5678', '서울특별시 강남구 역삼동', NOW()),
('이영희', 28, '010-2345-6789', '부산광역시 해운대구 우동', NOW()),
('박민수', 45, '010-3456-7890', '대구광역시 수성구 범어동', NOW()),
('최지현', 37, '010-4567-8901', '인천광역시 남동구 구월동', NOW()),
('정우성', 29, '010-5678-9012', '광주광역시 서구 치평동', NOW()),
('한지민', 41, '010-6789-0123', '대전광역시 유성구 봉명동', NOW()),
('오세훈', 35, '010-7890-1234', '울산광역시 남구 삼산동', NOW()),
('윤아름', 26, '010-8901-2345', '경기도 성남시 분당구 정자동', NOW()),
('강호동', 50, '010-9012-3456', '경상남도 창원시 의창구 팔용동', NOW()),
('배수지', 31, '010-0123-4567', '강원특별자치도 원주시 무실동', NOW());
