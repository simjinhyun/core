-- CREATE (Insert)
DROP PROCEDURE IF EXISTS `SPI_address_book`;
DELIMITER //
CREATE PROCEDURE `SPI_address_book` (
    p_name    VARCHAR(50),
    p_age     INT,
    p_phone   VARCHAR(20),
    p_address VARCHAR(100)
)
COMMENT '
========================================================
SPI_address_book
목적: 주소록 레코드 생성 (INSERT)
비고: created_at은 NOW()로 기록
========================================================
'
BEGIN
    INSERT INTO address_book (name, age, phone, address, created_at)
    VALUES (p_name, p_age, p_phone, p_address, NOW());
END //
DELIMITER ;


-- READ (Select All)
DROP PROCEDURE IF EXISTS `SPS_address_book_all`;
DELIMITER //
CREATE PROCEDURE `SPS_address_book_all` ()
COMMENT '
========================================================
SPS_address_book_all
목적: 주소록 전체 조회 (SELECT ALL)
비고: id 오름차순으로 정렬
========================================================
'
BEGIN
    SELECT id, name, age, phone, address, created_at
    FROM address_book
    ORDER BY id;
END //
DELIMITER ;


-- UPDATE
DROP PROCEDURE IF EXISTS `SPU_address_book`;
DELIMITER //
CREATE PROCEDURE `SPU_address_book` (
    p_id      INT,
    p_name    VARCHAR(50),
    p_age     INT,
    p_phone   VARCHAR(20),
    p_address VARCHAR(100)
)
COMMENT '
========================================================
SPU_address_book
목적: 주소록 레코드 수정 (UPDATE)
비고: id 기준으로 모든 필드 업데이트
========================================================
'
BEGIN
    UPDATE address_book
    SET name = p_name,
        age = p_age,
        phone = p_phone,
        address = p_address
    WHERE id = p_id;
END //
DELIMITER ;


-- DELETE
DROP PROCEDURE IF EXISTS `SPD_address_book`;
DELIMITER //
CREATE PROCEDURE `SPD_address_book` (
    p_id INT
)
COMMENT '
========================================================
SPD_address_book
목적: 주소록 레코드 삭제 (DELETE)
비고: id 기준으로 삭제
========================================================
'
BEGIN
    DELETE FROM address_book
    WHERE id = p_id;
END //
DELIMITER ;
